package main

import "sync"

const DEF_GOROUTINES = 1

type CDMConfig struct {
	downloadURL 	string
	goRoutines 		int
	contentMap		map[int][]byte
	*sync.Mutex
}

func NewCDM(
	downloadURL 	string,
	goRoutines		int,
) (*CDMConfig, error) {

	// set default number of goroutines 
	if goRoutines == 0 {
		goRoutines = DEF_GOROUTINES
	}

	contentMap := make(map[int][]byte)

	return &CDMConfig{
		downloadURL,
		goRoutines,
		contentMap,
		&sync.Mutex{},
	}, nil
}