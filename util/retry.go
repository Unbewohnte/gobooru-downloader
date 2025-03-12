package util

import (
	"fmt"
	"net/http"
	"time"
)

// DoGETRetry performs a GET request to the specified URL with retries.
// It retries the request up to `retries` times if it fails due to transient errors.
func DoGETRetry(client *http.Client, url string, retries uint) (*http.Response, error) {
	var response *http.Response
	var err error

	for attempt := uint(0); attempt <= retries; attempt++ {
		// Perform the GET request
		response, err = client.Get(url)
		if err == nil && response.StatusCode < 500 {
			return response, nil
		}

		// Wait before retrying
		if uint(attempt) != retries {
			time.Sleep(time.Duration(attempt*attempt) * time.Second)
		}
	}

	// Return the last error or response after all retries
	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("status code: %d", response.StatusCode)
}
