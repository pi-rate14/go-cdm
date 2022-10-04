package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {

	start := time.Now()

	goRoutines := flag.Int("t", 1, "number of goroutines")

	flag.Parse()

	if len(flag.Args()) == 0 {
		err := fmt.Errorf("enter download URL")
		log.Fatal(err)
	}

	downloadURL := flag.Args()[0]

	CDM, err := NewCDM(downloadURL, *goRoutines)
	if err != nil {
		log.Fatal(err)
	}

	ok, contentSize, err := CDM.acceptsMultiple()
	if err != nil {
		log.Fatal(err)
	}

	if !ok {
		fmt.Println("does not accept concurrent downloads, downloading using 1 goroutine")
	} else {
		CDM.downloadConcurrent(contentSize)
		fmt.Printf("accepts multiple, content length: %d", contentSize)

	}
	
	if err != nil {	
		log.Fatal(err)
	}

	log.Printf("main, execution time %s\n", time.Since(start))
}