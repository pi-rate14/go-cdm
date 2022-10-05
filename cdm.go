/*
	All the controllers for CDM
*/

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
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

// function to download a file that does not accept range header and uses 1 goroutine
func (CDM *CDMConfig) downloadSingle(contentSize int) error {
	CDM.contentMap[0] = CDM.outputFile
	CDM.downloadPart(nil, 0, 0, 0, CDM.outputFile)
	return CDM.appErr
}

// function to distribute chunks between goroutines
func (CDM *CDMConfig) downloadConcurrent(contentSize int) error {

	// close output file after all chunks are combined
	defer CDM.outputFile.Close()

	partSize := contentSize / CDM.goRoutines
	routine := 0
	wg := &sync.WaitGroup{}

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

		// update progress
		CDM.progress[routine] = NewProgressBar(0, tempLength-seek)

		// increment waitgroup before invoking goroutine
		wg.Add(1)

		// download chunk data
		go CDM.downloadPart(wg, routine, seek, tempLength, tempFile)

		// move onto the next goroutine
		routine++
	}

	// initialise the progress
	termSig := make(chan struct{})
	go CDM.initProgress(termSig)

	// wait for all goroutines to finish
	wg.Wait()
	termSig <- struct{}{}

	// check for errors
	if CDM.appErr != nil {
		os.Remove(CDM.outputFile.Name())
		return CDM.appErr
	}

	// join all the chunks
	return CDM.joinChunks()
}

// function to download chunk data for a goroutine
func (CDM *CDMConfig) downloadPart(
	wg *sync.WaitGroup,
	routine int,
	seek int,
	tempLength int,
	tempFile *os.File,
) {
	// decrement waitGroup once this download is complete
	defer wg.Done()
	// TODO : strings.NewReader
	req, err := http.NewRequest(http.MethodGet, CDM.downloadURL, nil)
	if err != nil {
		CDM.appErr = err
		return
	}

	// if a range is specified, modify the header
	if seek+tempLength != 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", seek, tempLength))
	}

	// make a request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		CDM.appErr = err
		return
	}
	defer res.Body.Close()

	// create a read buffer to write response body to progress bar
	readBuffer := make([]byte, READ_BUFFER_SIZE)
	var totalRead int

	for {
		select {
		case sigErr := <-CDM.termErr:
			CDM.appErr = sigErr
			return
		default:
			err = CDM.readResponseBody(res, readBuffer, tempFile, &totalRead, routine)

			// if we reach the end of the temporary file, return from this goroutine
			if err == io.EOF {
				return
			}
			if err != nil {
				CDM.appErr = err
				return
			}
		}

	}
}

// funtion to read response body and update progress bar
func (CDM *CDMConfig) readResponseBody(res *http.Response, readBuffer []byte, tempFile io.Writer, totalRead *int, routine int) error {

	// read from response body and write to temporary file
	readBytes, err := res.Body.Read(readBuffer)
	if readBytes > 0 {
		tempFile.Write(readBuffer[:readBytes])
	}

	if err != nil {
		return err
	}

	// update the total bytes read and the progress bar
	*totalRead += readBytes

	CDM.mutex.Lock()
	CDM.progress[routine].current = *totalRead
	CDM.mutex.Unlock()

	return nil
}

// function to join the temp files after they are downloaded
func (CDM *CDMConfig) joinChunks() error {

	// for each goroutine, copy the file data in output file, in order
	for routine := 0; routine < len(CDM.contentMap); routine++ {

		tempFile := CDM.contentMap[routine]
		tempFile.Seek(0, 0)
		_, err := io.Copy(CDM.outputFile, tempFile)
		if err != nil {
			return err
		}

	}

	// find the size of the output file
	oFile, err := CDM.outputFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	bytesWritten := oFile.Size()

	log.Printf("Download complete. Output size: %d\n", bytesWritten)

	return nil
}

func (CDM *CDMConfig) handleSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		sig := <-signalChan
		for routine := 0; routine < len(CDM.contentMap); routine++ {
			CDM.termErr <- fmt.Errorf("user stopped : %v", sig)
		}
	}()
}
