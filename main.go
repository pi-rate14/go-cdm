package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
)

func main() {

	goRoutines := flag.Int("t", 1, "number of goroutines")

	flag.Parse()

	if len(flag.Args()) == 0 {
		err := fmt.Errorf("enter download URL")
		log.Fatal(err)
	}

	downloadURL, err := url.ParseRequestURI(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}

	CDM, err := NewCDM(downloadURL, *goRoutines)
	if err != nil {
		log.Fatal(err)
	}

	ok, contentLength, err := CDM.acceptsMultiple()
	if err != nil {
		log.Fatal(err)
	}

	if !ok {
		fmt.Println("does not accept concurrent downloads, downloading using 1 goroutine")
	} else {
		// CDM.downloadConcurrent()
		fmt.Printf("accepts multiple, content length: %d", contentLength)

	}
	
	if err != nil {	
		log.Fatal(err)
	}

}