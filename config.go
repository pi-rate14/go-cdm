/*
	Config file to configure a new CDM
*/
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	DEF_GOROUTINES   = 1            // deafult goroutines if not specified in terminal
	MAX_GOROUTINES   = 30           // max goroutines possible
	READ_BUFFER_SIZE = 500          // size of the buffer that reads response for each goroutine
	TEMP_FILE_NAME   = "temp.chunk" // name of the temporary file used to store chunk data
	PROGRESS_SIZE    = 100          // size in characters of the progress bar
)

type CDMConfig struct {
	downloadURL    string               // download link of the file
	goRoutines     int                  // no of cocurrent download connections
	contentMap     map[int]*os.File     // map goroutine to corresponding temporary file
	outputFile     *os.File             // final output file
	outputFileName string               // name of the output file
	termErr        chan error           // receive signals from the terminal during download
	appErr         error                // error received from goroutines
	progress       map[int]*progressBar // map goroutine to its progress
	mutex          *sync.RWMutex        // mutex to handle concurrent writes and reads on the maps
}

func NewCDM(
	downloadURL string,
	goRoutines int,
) (*CDMConfig, error) {

	// set default number of goroutines if not received from user
	if goRoutines == 0 {
		goRoutines = DEF_GOROUTINES
	}

	// set max goroutines if they exceed the range
	if goRoutines > MAX_GOROUTINES {
		goRoutines = MAX_GOROUTINES
	}

	contentMap := make(map[int]*os.File)

	outputFile, err := setOutputFile(&downloadURL)
	if err != nil {
		log.Fatal(err)
	}

	termErr := make(chan error)

	progress := make(map[int]*progressBar)

	return &CDMConfig{
		downloadURL:    downloadURL,
		goRoutines:     goRoutines,
		contentMap:     contentMap,
		outputFile:     outputFile,
		outputFileName: filepath.Base(downloadURL),
		termErr:        termErr,
		appErr:         nil,
		progress:       progress,
		mutex:          &sync.RWMutex{},
	}, nil
}

// TODO: dissolve this function
func setOutputFile(downloadURL *string) (*os.File, error) {
	// get the current working directory
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// make a filename for the output file from the download URL
	oFileName := fmt.Sprintf("%s/%s", currentDirectory, filepath.Base(*downloadURL))

	// check if a file with the same name already exists
	_, err = os.Stat(oFileName)
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("file already exists")
	}

	// create the output file and give it read/write/append permissions
	output, err := os.OpenFile(oFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return nil, fmt.Errorf("error while creating file : %v", err)
	}

	return output, nil
}
