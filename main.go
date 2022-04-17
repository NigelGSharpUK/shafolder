package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	// Deal with command line arguments
	flagTerse := flag.Bool("terse", false, "Shows fewer helpful hints")
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatal("Please provide one filename")
	}
	flagFilename := flag.Args()[0]

	// Open the file
	f, err := os.Open(flagFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Calculate the sha256 of file contents
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	// Output
	if !*flagTerse {
		fmt.Print("sha256 of the file " + flagFilename + ": ")
	}
	fmt.Printf("%x", h.Sum(nil))
}
