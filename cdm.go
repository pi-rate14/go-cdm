package main

import "fmt"

const DEF_GOROUTINES = 4

type Config struct {
	downloadURL 	string
	goRoutines 		int
}

func CDM(
	downloadURL 	string,
	goRoutines		int,
) (*Config, error) {
	
	if len(downloadURL) == 0 {
		err := fmt.Errorf("please enter a download URL")
		return nil, err
	}

	if goRoutines == 0 {
		goRoutines = DEF_GOROUTINES
	}

	return &Config{
		downloadURL,
		goRoutines,
	}, nil
}