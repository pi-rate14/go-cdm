package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// Function that checks if the request has Accept-Range header and returns content length
func (CDM *CDMConfig) acceptsMultiple() (bool, int, error) {

	URL := CDM.downloadURL

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

func (CDM *CDMConfig) downloadConcurrent(contentSize int) error {

	defer CDM.outputFile.Close()

	partSize := contentSize / CDM.goRoutines
	routine := 0
	var wg sync.WaitGroup

	errs := make(chan error)

	for seek := 0; seek < contentSize; seek += partSize + 1 {

		// defining the range for current download
		tempLength := seek + partSize

		// checking if the last chunk size exceeds total content size
		if contentSize < tempLength {
			tempLength = contentSize
		}

		// create temp file to write chunk data
		tempFile, err := os.CreateTemp("", TEMP_FILE_NAME)
		if err != nil {
			return err
		}
		defer tempFile.Close()
		defer os.Remove(tempFile.Name())

		// map content range to goroutine
		CDM.contentMap[routine] = tempFile

		// increment waitgroup before invoking goroutine
		wg.Add(1)

		// download chunk data
		go CDM.downloadPart(&wg, routine, seek, tempLength, tempFile, errs)

		// move onto the next goroutine
		routine++
	}

	// wait for all goroutines to finish
	wg.Wait()
	// check for errors
	return CDM.joinChunks()
}

func (CDM *CDMConfig) downloadPart(
	wg *sync.WaitGroup,
	routine int,
	seek int,
	tempLength int,
	tempFile *os.File,
	errschan chan error,

) {
	// decrement waitGroup once this download is complete
	defer wg.Done()

	fmt.Printf("Downloading on goroutine #%d\n", routine)

	req, err := http.NewRequest(http.MethodGet, CDM.downloadURL, nil)
	if err != nil {
		log.Fatal(err)
		errschan <- err
	}

	if seek+tempLength != 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", seek, tempLength))
	}

	// make a request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		errschan <- err
	}
	defer res.Body.Close()

	_, err = io.Copy(tempFile, res.Body)
	if err != nil {
		log.Fatal(err)
		errschan <- err
	}

	// CDM.mutex
	// CDM.contentMap[routine] = append(CDM.contentMap[routine], data...)
	// CDM.Unlock()

}

func (CDM *CDMConfig) joinChunks() error {

	for routine := 0; routine < len(CDM.contentMap); routine++ {

		tempFile := CDM.contentMap[routine]
		tempFile.Seek(0, 0)
		_, err := io.Copy(CDM.outputFile, tempFile)
		if err != nil {
			return err
		}

	}

	return nil
}
