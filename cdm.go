package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
)

// Function that checks if the request has Accept-Range header and returns content length
func (CDM *CDMConfig) acceptsMultiple() (bool, int, error) {
	
	URL := CDM.downloadURL.String()

	resp, err := http.Head(URL)
	if err != nil {
		return false, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, 0, fmt.Errorf("request unsuccessful")
	}

	cl := resp.Header.Get("Content-Length")
	contentLength, err := strconv.Atoi(cl) 
	if err != nil {
		return false, 0, err
	}

	if resp.Header.Get("Accept-Ranges") != "bytes" {
		return false, contentLength, nil
	}

	return true, contentLength, nil
} 

func (CDM *CDMConfig) downloadConcurrent(contentSize int) {
	
	chunkSize := contentSize / CDM.goRoutines
	start := 0
	var wg &sync.WaitGroup

	for i:=0; i<contentSize; i++ {
		// defining the range for current download
		tempLength := i + chunkSize
		// checking if the last chunk size exceeds total content size
		tempLength = math.MinInt64(tempLength, contentSize)


	}
}