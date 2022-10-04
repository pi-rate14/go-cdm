package main

import (
	"net/url"
)

const DEF_GOROUTINES = 1

type CDMConfig struct {
	downloadURL 	*url.URL
	goRoutines 		int
}

func NewCDM(
	downloadURL 	*url.URL,
	goRoutines		int,
) (*CDMConfig, error) {

	// set default number of goroutines 
	if goRoutines == 0 {
		goRoutines = DEF_GOROUTINES
	}

	return &CDMConfig{
		downloadURL,
		goRoutines,
	}, nil
}