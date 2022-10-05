package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	DEF_GOROUTINES = 1
	MAX_GOROUTINES = 30
	TEMP_FILE_NAME = "temp.chunk"
)

type CDMConfig struct {
	downloadURL    string
	goRoutines     int
	contentMap     map[int]*os.File
	outputFile     *os.File
	outputFileName string
	mutex          *sync.RWMutex
}

func NewCDM(
	downloadURL string,
	goRoutines int,
) (*CDMConfig, error) {

	// set default number of goroutines
	if goRoutines == 0 {
		goRoutines = DEF_GOROUTINES
	}

	if goRoutines > MAX_GOROUTINES {
		goRoutines = MAX_GOROUTINES
	}

	contentMap := make(map[int]*os.File)

	outputFile, err := setOutputFile(downloadURL)
	if err != nil {
		log.Fatal(err)
	}

	return &CDMConfig{
		downloadURL:    downloadURL,
		goRoutines:     goRoutines,
		contentMap:     contentMap,
		outputFile:     outputFile,
		outputFileName: filepath.Base(downloadURL),
		mutex:          &sync.RWMutex{},
	}, nil
}

func setOutputFile(downloadURL string) (*os.File, error) {
	// get the current working directory
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	oFileName := fmt.Sprintf("%s/%s", currentDirectory, filepath.Base(downloadURL))
	_, err = os.Stat(oFileName)
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("file already exists")
	}

	output, err := os.Create(oFileName)
	if err != nil {
		return nil, err
	}

	return output, nil
}
