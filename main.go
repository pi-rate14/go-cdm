package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {

	// start a timer to find total duration of download
	start := time.Now()

	// get goroutines from terminal input
	goRoutines := flag.Int("t", 1, "number of goroutines")

	flag.Parse()

	// check if download URL provided
	if len(flag.Args()) == 0 {
		err := fmt.Errorf("enter download URL")
		log.Fatal(err)
	}

	// parse download URL
	downloadURL := flag.Args()[0]

	// get a new CDM inntance
	CDM, err := NewCDM(downloadURL, *goRoutines)
	if err != nil {
		log.Fatal(err)
	}

	// check if download URL provides range downloads
	ok, contentSize, err := CDM.acceptsMultiple()
	if err != nil {
		log.Fatal(err)
	}

	// download concurrently or on 1 goroutine based on above result
	if !ok {
		fmt.Printf("does not accept concurrent downloads, downloading using 1 goroutine\n")
		err = CDM.downloadSingle(contentSize)
	} else {
		fmt.Printf("accepts multiple, downloading using %d goroutines", CDM.goRoutines)
		err = CDM.downloadConcurrent(contentSize)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Printf(" Initial size : %d \n Time Taken : %s\n", contentSize, time.Since(start))
}
