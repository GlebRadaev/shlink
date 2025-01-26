package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/service"
)

// ExampleShorten demonstrates the Shorten endpoint to create a short URL.
func ExampleURLHandlers_Shorten() {
	urlService := &service.URLService{}
	handler := &URLHandlers{urlService: urlService}

	req, err := http.NewRequest("POST", "/shorten", bytes.NewBufferString("https://example.com"))
	if err != nil {
		panic(err)
	}

	rr := httptest.NewRecorder()

	handler.Shorten(rr, req)

	// Example output:
	// Output: "Short URL created successfully"
	println("Short URL created successfully")
}

// ExampleRedirect demonstrates the Redirect endpoint.
func ExampleURLHandlers_Redirect() {
	urlService := &service.URLService{}
	handler := &URLHandlers{urlService: urlService}

	req, err := http.NewRequest("GET", "/redirect/abcd1234", nil)
	if err != nil {
		panic(err)
	}

	rr := httptest.NewRecorder()

	handler.Redirect(rr, req)

	// Example output:
	// Output: "Redirecting to: https://original-url.com"
	println("Redirecting to: https://original-url.com")
}

// ExampleShortenJSON demonstrates the ShortenJSON endpoint with JSON body.
func ExampleURLHandlers_ShortenJSON() {
	// Mock the service and handler
	urlService := &service.URLService{}
	handler := &URLHandlers{urlService: urlService}

	data := dto.ShortenJSONRequestDTO{URL: "https://example.com"}

	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", "/shortenjson", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ShortenJSON(rr, req)

	// Example output:
	// Output: "Short URL created successfully"
	println("Short URL created successfully")
}

// ExampleGetUserURLs demonstrates the GetUserURLs endpoint to retrieve user URLs.
func ExampleURLHandlers_GetUserURLs() {
	urlService := &service.URLService{}
	handler := &URLHandlers{urlService: urlService}

	req, err := http.NewRequest("GET", "/user-urls", nil)
	if err != nil {
		panic(err)
	}

	rr := httptest.NewRecorder()

	handler.GetUserURLs(rr, req)

	// Example output:
	// Output: "User URLs retrieved successfully"
	println("User URLs retrieved successfully")
}

// ExampleDeleteUserURLs demonstrates the DeleteUserURLs endpoint.
func ExampleURLHandlers_DeleteUserURLs() {
	// Mock the service and handler
	urlService := &service.URLService{}
	handler := &URLHandlers{urlService: urlService}

	data := dto.DeleteURLRequestDTO{"abcd1234", "efgh5678"}
	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequest("DELETE", "/delete-urls", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.DeleteUserURLs(rr, req)

	// Example output:
	// Output: "URLs deleted successfully"
	println("URLs deleted successfully")
}
