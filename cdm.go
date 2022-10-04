package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	
	partSize := contentSize / CDM.goRoutines
	routine := 0
	var wg sync.WaitGroup

	errs := make(chan error)
	
	for seek:=0; seek<contentSize; seek += partSize + 1 {

		// defining the range for current download
		tempLength := seek + partSize

		// checking if the last chunk size exceeds total content size
		if contentSize < tempLength {
			tempLength = contentSize
		}

		// map content range to goroutine
		CDM.contentMap[routine] =  make([]byte, 0)

		// increment waitgroup before invoking goroutine
		wg.Add(1)

		// download chunk data
		go CDM.downloadPart(&wg, routine, seek, tempLength, errs)

		// move onto the next goroutine
		routine++
	}

	// wait for all goroutines to finish
	wg.Wait()	
	// check for errors
	return CDM.joinChunks()
}

func (CDM *CDMConfig) downloadPart(
	wg 				*sync.WaitGroup,
	routine 		int,
	seek			int,
	tempLength		int,
	errschan		chan error,

) {
	// decrement waitGroup once this download is complete
	defer wg.Done()

	fmt.Printf("Downloading on goroutine #%d", routine)

	req, err := http.NewRequest(http.MethodGet, CDM.downloadURL, nil)
	if err != nil {
		log.Fatal(err)
		errschan <- err
	}

	if(seek+tempLength != 0) {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", seek, tempLength))
	}

	// make a request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		errschan <- err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		errschan <- err
	}

	CDM.Lock()
	CDM.contentMap[routine] = append(CDM.contentMap[routine], data...)
	CDM.Unlock()
	
}

func (CDM *CDMConfig) joinChunks() error {

	// get the current working directory	
	currentDirectory, err := os.Getwd()
	if err != nil {
		return err
	}

	oFileName := fmt.Sprintf("%s/%s", currentDirectory, filepath.Base(CDM.downloadURL))

	output, err := os.Create(oFileName)
	if err != nil {
		return err
	}
	defer output.Close()

	buffer := bytes.NewBuffer(nil)
	for i := 0; i < len(CDM.contentMap); i++ {

		buffer.Write(CDM.contentMap[i])
	
	} 
	_, err = buffer.WriteTo(output)
	if err != nil {
		return err	
	}

	return nil
}