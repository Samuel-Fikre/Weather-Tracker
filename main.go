package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type apiConfigData struct {
	OpenWeatherMapApiKey string `json:"OpenWeatherMapApiKey"`
}

// This defines a new data structure that will hold the weather information retrieved from an API response.

// nested struct
type weatherData struct {
	Name string `json:"name"`
	Main struct {
		Kelvin float64 `json:"temp"`
	} `json:"main"`
}

// filename string: The function accepts the name of the file (as a string) that contains the API configuration.

// (apiConfigData, error): This tells you what the magic box will give you back when it finishes the task. Its like saying, "If I read the book correctly, I will give you two things: the secrets (called apiConfigData), or if I mess up, Ill give you a message saying what went wrong (an error).

// So, imagine you give the box a book called config.json.If the box reads the book and finds the secrets inside (API key, etc.), it will give those to you.

func loadApiConfig(filename string) (apiConfigData, error) {

	// os.ReadFile(filename):

	// This is a built-in function that reads the contents of a file specified by filename.
	bytes, err := os.ReadFile(filename)

	// apiConfigData{}: This is an empty value of the apiConfigData struct, indicating that no valid configuration data could be loaded. The {} creates a new, empty instance of the struct.

	if err != nil {
		return apiConfigData{}, err
	}

	var c apiConfigData

	// json.Unmarshal is a function that takes a JSON-encoded byte slice (in this case, bytes) and decodes (unmarshals) it into the provided Go variable (in this case, &c).

	//bytes: This is the byte slice containing the raw JSON data that was read from the file.
	//&c: The & symbol indicates you are passing a pointer to c. This means json.Unmarshal will directly modify the contents of c as it parses the JSON into the corresponding fields of the apiConfigData struct

	// err: This captures any error that might occur during the unmarshalling process. If the JSON is malformed or doesn't match the structure of apiConfigData, an error will be returned.
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return apiConfigData{}, err
	}
	//If no error occurred (meaning err == nil), the function proceeds to this line.
	return c, nil
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from go!\n"))
}

func query(city string) (weatherData, error) {

	// load the .env file
	apiConfig, err := loadApiConfig(".apiConfig")
	if err != nil {
		return weatherData{}, err
	}
	// .env file loaded
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", city, apiConfig.OpenWeatherMapApiKey)

	fmt.Println("Requesting URL:", url)

	resp, err := http.Get(url)

	if err != nil {
		return weatherData{}, err
	}

	// Purpose: When you make an HTTP request using http.Get, it returns a response (resp) that contains a body (resp.Body). This body needs to be closed once you are done with it to free up network resources and avoid memory leaks.

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body) // Updated from ioutil to io
	if err != nil {
		return weatherData{}, err
	}
	fmt.Println(string(body)) // Log the raw response body for debugging

	var d weatherData
	if err := json.Unmarshal(body, &d); err != nil {
		return weatherData{}, err
	}

	return d, nil

}

func main() {
	http.HandleFunc("/hello", hello)

	// URL.Path is a string that contains the path part of the URL (the part after the domain name). For example, if the full URL is http://example.com/cities/new-york, e.URL.Path would be /cities/new-york
	// strings.SplitN is a function that splits a string into a slice of substrings, using the specified delimiter (in this case, the slash /). The N in SplitN means that it will split the string into at most N parts. Here, N is 3.
	//So, for a URL path like /cities/new-york, this would split the string at each /, but only up to 3 pieces:

	//Part 1: "" (an empty string because the path starts with a /)
	//Part 2: "cities"
	//Part 3: "new-york"

	//The slice indexing [2] accesses the third element in the slice created by strings.SplitN.

	http.HandleFunc("/weather/",
		func(w http.ResponseWriter, r *http.Request) {
			parts := strings.SplitN(r.URL.Path, "/", 3)

			if len(parts) < 3 {
				http.Error(w, "City not specified", http.StatusBadRequest)
				return
			}
			city := parts[2]
			data, err := query(city)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError) // Fixed Error() to err.Error()
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf8")
			json.NewEncoder(w).Encode(data)
		}) // Ensure this closing brace is correctly placed

	http.ListenAndServe(":8080", nil) // Ensure the port is specified correctly as a string
}
