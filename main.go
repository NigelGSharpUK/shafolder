package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/tyler-smith/go-bip39"
)

func main() {
	// Deal with command line arguments
	flagTerse := flag.Bool("terse", false, "Shows fewer helpful hints")
	flagBip39 := flag.Bool("bip39", false, "Shows BIP39 mnenomic instead of sha256")
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatal("Please provide one filename")
	}
	flagFilename := flag.Args()[0]

	// Is it a folder?
	fileInfo, err := os.Stat(flagFilename)
	if err != nil {
		log.Fatal(err)
	}
	if fileInfo.IsDir() {
		log.Fatal("That's a folder dumbass!")
	}

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
	if *flagBip39 {
		m, err := bip39.NewMnemonic(h.Sum(nil))
		if err != nil {
			log.Fatal(err)
		}
		if !*flagTerse {
			fmt.Println("BIP39 mnenomic of the sha256 hash of file " + flagFilename + ": ")
		}
		fmt.Println(m)
	} else {
		if !*flagTerse {
			fmt.Print("sha256 hash of the file " + flagFilename + ": ")
		}
		fmt.Printf("%x\n", h.Sum(nil))
	}
}
