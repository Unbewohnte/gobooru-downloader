/*
   gobooru-downloader
   Copyright (C) 2025 Kasyanov Nikolay Alexeevich (Unbewohnte)

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package proxy

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var MAXRETRIES uint = 5

// Perform a GET request to the specified URL with retries.
// It retries the request up to `retries` times if it fails due to transient errors.
func DoGETRetry(client *http.Client, url string) (*http.Response, error) {
	var response *http.Response
	var err error

	for attempt := uint(0); attempt <= MAXRETRIES; attempt++ {
		// Perform the GET request
		response, err = client.Get(url)
		if err == nil && response.StatusCode < 500 {
			return response, nil
		}

		// Wait before retrying
		if attempt < MAXRETRIES {
			time.Sleep(time.Duration(attempt*attempt) * time.Second)
		}
	}

	// Return the last error or response after all retries
	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("status code: %d", response.StatusCode)
}

// Downloads a content from the given URL and returns its content as a byte slice
func GetContents(client *http.Client, contentURL string) ([]byte, error) {
	response, err := DoGETRetry(client, contentURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", response.StatusCode)
	}

	// Read the content into a byte slice
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
