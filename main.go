package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

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

	// Walk all the files/folders from the root folder (or just file!) named flagFilename
	err := filepath.Walk(flagFilename, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if info.IsDir() {
			if !*flagTerse {
				fmt.Println("Folder: ", path)
			}
		} else {
			if !*flagTerse {
				fmt.Println("File: ", path)
			}

			// Open the file
			f, err := os.Open(path)
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
					fmt.Print("sha256 hash of the file: ")
				}
				fmt.Printf("%x\n", h.Sum(nil))
			}
		}
		return nil
	})

	// Did the walk fail?
	if err != nil {
		log.Fatal(err)
	}
}
