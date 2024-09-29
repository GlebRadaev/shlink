package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidURL(t *testing.T) {
    testCases := []struct {
        name       string
        url        string
        isValid    bool
    }{
        {"Valid HTTP URL", "http://example.com", true},
        {"Valid HTTPS URL", "https://example.com", true},
        {"Invalid FTP URL", "ftp://example.com", false},
        {"URL with space", "http://example .com", false},
        {"URL with hash symbol", "http://example.com#", false},
        {"Empty URL after protocol", "https://", false},
        {"URL without domain", "http://example", false},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := isValidURL(tc.url)
            assert.Equal(t, tc.isValid, result)
        })
    }
}

func TestShortenURL(t *testing.T) {
    tests := []struct {
        name           string
        contentType    string
        body           string
        expectedCode   int
        expectedError  string
    }{
        {
            name:          "Invalid Content-Type",
            contentType:   "application/json",
            body:          "http://example.com",
            expectedCode:  http.StatusBadRequest,
            expectedError: invalidContentTypeMsg,
        },
        {
            name:          "Invalid URL Format",
            contentType:   "text/plain",
            body:          "http://example/com",
            expectedCode:  http.StatusBadRequest,
            expectedError: invalidURLFormatMsg,
        },
        {
            name:          "URL Too Long",
            contentType:   "text/plain",
            body:          "http://."+strings.Repeat("a", 2049),
            expectedCode:  http.StatusBadRequest,
            expectedError: urlTooLongMessage,
        },
        {
            name:         "Successful URL Shortening",
            contentType:  "text/plain",
            body:         "http://example.com",
            expectedCode: http.StatusCreated,
        },
    }   
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            shortUrls = make(map[string]string) 

            req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
            req.Header.Set("Content-Type", tt.contentType)
            rr := httptest.NewRecorder()

            shortenURL(rr, req)

            assert.Equal(t, tt.expectedCode, rr.Code)   
            if tt.expectedError != "" {
                assert.Equal(t, tt.expectedError, strings.TrimSpace(rr.Body.String()))
            } else {
                body := rr.Body.String()
                expectedPrefix := fmt.Sprintf("%s:%s/", baseURL, port)
                assert.Contains(t, body, expectedPrefix)
            
                parts := strings.Split(strings.TrimSpace(body), "/")
                generatedID := parts[len(parts)-1]
            
                savedURL, ok := shortUrls[generatedID]
                assert.True(t, ok, "Expected ID not found in shortUrls map")
                assert.Equal(t, "http://example.com", savedURL)
            }
        })
    }
}

func TestRedirectURL(t *testing.T) {
    tests := []struct {
        name          string
        id            string
        setupShortURL func()
        expectedCode  int
        expectedError string
        expectedURL   string
    }{
        {
            name:          "Valid ID",
            id:            "12345678",
            setupShortURL: func() {
            	shortUrls["12345678"] = "http://example.com"
            },
            expectedCode:  http.StatusTemporaryRedirect,
            expectedURL:   "http://example.com",
        },
        {
            name:          "Invalid ID",
            id:            "12345678",
            setupShortURL: func() {},
            expectedCode:  http.StatusBadRequest,
            expectedError: invalidIDMsg,
        },
        {
            name:          "Invalid ID Length",
            id:            "short",
            setupShortURL: func() {},
            expectedCode:  http.StatusBadRequest,
            expectedError: invalidIDLengthMsg,
        },
    }   
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            shortUrls = make(map[string]string)
            tt.setupShortURL()

            req := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)
            rr := httptest.NewRecorder()
            
            redirectURL(rr, req)

            assert.Equal(t, tt.expectedCode, rr.Code)
            if tt.expectedError != "" {
                assert.Equal(t, tt.expectedError, strings.TrimSpace(rr.Body.String()))
            } else {
                assert.Equal(t, tt.expectedURL, rr.Header().Get("Location"))
            }
        })
    }
}

func TestHandleRequest(t *testing.T) {
    testCases := []struct {
        name           string
        method         string
        url            string
        body           string
        contentType    string
        expectedStatus int
        expectedBody   string
        expectedHeader string
    }{
        {
            name:           "POST method - shorten URL",
            method:         http.MethodPost,
            url:            "/",
            body:           "http://example.com",
            contentType:    "text/plain",
            expectedStatus: http.StatusCreated,
            expectedBody:   baseURL,
        },
        {
            name:           "GET method - redirect URL",
            method:         http.MethodGet,
            url:            "/abc12345",
            expectedStatus: http.StatusTemporaryRedirect,
            expectedHeader: "http://example.com",
        },
        {
            name:           "Invalid method - Method not allowed",
            method:         http.MethodDelete,
            url:            "/",
            expectedStatus: http.StatusBadRequest,
            expectedBody:   "Method not allowed",
        },
    }
    shortUrls["abc12345"] = "http://example.com"

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            var req *http.Request
            if tc.body != "" {
                req = httptest.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
            } else {
                req = httptest.NewRequest(tc.method, tc.url, nil)
            }
            if tc.contentType != "" {
                req.Header.Set("Content-Type", tc.contentType)
            }

            rr := httptest.NewRecorder()
            handleRequest(rr, req)
            result := rr.Result()
            defer result.Body.Close()

            assert.Equal(t, tc.expectedStatus, result.StatusCode)
            if tc.expectedBody != "" {
                body, _ := io.ReadAll(result.Body)
                assert.Contains(t, string(body), tc.expectedBody)
            }
            if tc.expectedHeader != "" {
                location := result.Header.Get("Location")
                assert.Equal(t, tc.expectedHeader, location)
            }
        })
    }
}