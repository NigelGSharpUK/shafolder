package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

func main() {
	// Deal with command line arguments
	flagTerse := flag.Bool("terse", false, "Shows fewer helpful hints")
	flagBip39 := flag.Bool("bip39", false, "Shows BIP39 mnenomic instead of sha256")
	flagRename := flag.Bool("rename", false, "Copy each file prepending two BIP39 words to filename")
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatal("Please provide one filename")
	}
	flagFilename := flag.Args()[0]
	if *flagRename && !*flagBip39 {
		log.Fatal("-rename may only be used in conjunction with -bip39")
	}

	// Walk all the files/folders from the root folder (or just file!) named flagFilename.
	// Put each file's hash into hashArray.
	var hashArray [][]byte
	currentFolder := ""
	err := filepath.Walk(flagFilename, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if info.IsDir() {
			if !*flagTerse {
				fmt.Println("Folder: ", path)
			}
			currentFolder = path
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

			// Put the sha256 hash of the file into hashArray
			hashArray = append(hashArray, h.Sum(nil))

			m := ""

			// Output
			if *flagBip39 {
				m, err = bip39.NewMnemonic(h.Sum(nil))
				if err != nil {
					log.Fatal(err)
				}
				if !*flagTerse {
					fmt.Println("BIP39 mnenomic of the sha256 hash of file: ")
				}
				fmt.Println(m)

				// Copy file with mnemonic at front of filename
				if *flagRename {
					words := strings.Fields(m)
					oldFilename := info.Name()
					newFilename := words[0] + " " + words[1] + " " + oldFilename

					source, err := os.Open(currentFolder + "\\" + oldFilename)
					if err != nil {
						log.Fatal(err)
					}
					defer source.Close()

					destination, err := os.Create(currentFolder + "\\" + newFilename)
					if err != nil {
						log.Fatal(err)
					}
					defer destination.Close()
					_, err = io.Copy(destination, source)
					if err != nil {
						log.Fatal(err)
					}
				}
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

	fmt.Printf("There are %d files.\n", len(hashArray))

	// Sort the hashes (so that order is deterministically derived from contents, not filenames)
	sort.Slice(hashArray, func(i, j int) bool {
		return bytes.Compare(hashArray[i], hashArray[j]) == -1
	})

	// Concatenate the hashes
	var concatArray []byte
	for _, element := range hashArray {
		concatArray = append(concatArray, element...)
	}

	// Hash the hashes
	hash := sha256.Sum256(concatArray)

	// Output
	if *flagBip39 {
		m, err := bip39.NewMnemonic(hash[:])
		if err != nil {
			log.Fatal(err)
		}
		if !*flagTerse {
			fmt.Println("BIP39 mnenomic of ALL the file contents together:")
		}
		fmt.Println(m)
	} else {
		if !*flagTerse {
			fmt.Print("sha256 hash of ALL the file contents together:")
		}
		fmt.Printf("%x\n", hash)
	}
}
