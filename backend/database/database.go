package database

import (
	"errors"
	"io"
	"net/http"
	"os"
	"time"
)

var ErrFailedRequest = errors.New("Failed to build request")
var ErrFailedFetch = errors.New("Failed to fetch data")
var ErrFailedReadBody = errors.New("Failed to read response body")

func Request(httpMethod, query string, data io.Reader) (body []byte, err error) {
	// Build request
	API_URL := os.Getenv("DATABASE_API_URL")
	API_KEY := os.Getenv("DATABASE_API_KEY")
	if API_URL == "" || API_KEY == "" {
		return nil, ErrFailedRequest
	}
	request, err := http.NewRequest(httpMethod, API_URL+"/"+query, data)
	if err != nil {
		return nil, ErrFailedRequest
	}
	request.Header.Add("apikey", API_KEY)

	// Send request
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, ErrFailedFetch
	}
	defer response.Body.Close()

	// Read response body
	body, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, ErrFailedReadBody
	}

	return body, nil
}
