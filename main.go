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

	"github.com/iancoleman/strcase"
	"github.com/tyler-smith/go-bip39"
)

func main() {
	// Deal with command line arguments
	flagBip39 := flag.Bool("bip39", false, "Shows BIP39 mnenomic instead of sha256")
	flagVerbose := flag.Bool("verbose", false, "Show sha256 or mnenomic for every file in folder")
	flagNameFiles := flag.Bool("namefiles", false, "Copy each file, prepending two BIP39 words to filename, into \\.Bip39\\ folder")
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatal("Please provide one filename")
	}
	flagFilename := flag.Args()[0]
	if *flagNameFiles && !*flagBip39 {
		log.Fatal("-namefiles may only be used in conjunction with -bip39")
	}

	// Walk all the files/folders from the root folder (or just file!) named flagFilename.
	// Put each file's hash into hashArray.
	var hashArray [][]byte
	err := filepath.Walk(flagFilename, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if info.Name()[0:1] == "." {
			// Skip files and folders beginning with '.'
		} else if strings.Contains(path, "\\.") {
			// Skip files and folders with \. in the path (annoying gotcha for the above!)
		} else if info.IsDir() {
			// It's a folder
		} else {
			// It's a file
			if *flagVerbose {
				fmt.Println("FILE: ", path)
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

			// Output (for one of the files)
			if *flagBip39 {
				// Turn the 256 bit sha256 hash into a 24 word mnemonic
				m, err := bip39.NewMnemonic(h.Sum(nil))
				if err != nil {
					log.Fatal(err)
				}
				words := strings.Fields(m)

				// Print as two lines
				if *flagVerbose {
					fmt.Println("  " + strings.Join(words[0:12], " "))
					fmt.Println("  " + strings.Join(words[12:24], " "))
				}

				// Copy to file with mnemonic at front of filename in a folder \.Bip39\
				if *flagNameFiles {
					oldFilename := info.Name()
					newFilename := strcase.ToCamel(words[0]+" "+words[1]) + " " + oldFilename

					source, err := os.Open(path)
					if err != nil {
						log.Fatal(err)
					}
					defer source.Close()

					// Put it in a folder \.Bip39\
					bipFolder := filepath.Dir(path) + "\\.Bip39"
					err = os.MkdirAll(bipFolder, 0755)
					if err != nil {
						log.Fatal(err)
					}
					destination, err := os.Create(bipFolder + "\\" + newFilename)
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
				// sha256
				if *flagVerbose {
					fmt.Printf("  %x\n", h.Sum(nil))
				}
			}
		}
		return nil
	})

	// Did the walk fail?
	if err != nil {
		log.Fatal(err)
	}

	// Sort the hashes (so that order is deterministically derived from contents, not filenames)
	sort.Slice(hashArray, func(i, j int) bool {
		return bytes.Compare(hashArray[i], hashArray[j]) == -1
	})

	// Concatenate the hashes
	var concatArray []byte
	for _, element := range hashArray {
		concatArray = append(concatArray, element...)
	}

	// Decide on output (for all the files put together)
	var hash []byte
	if len(hashArray) > 1 {
		// Hash the hashes
		hashOfHashes := sha256.Sum256(concatArray)
		hash = hashOfHashes[:]

		// Print summary title
		if *flagVerbose {
			fmt.Printf("TOGETHER: (%d files)\n", len(hashArray))
		}
	} else {
		// The one and only, so don't hash the hash
		hash = hashArray[0]
	}

	// Print output (for all the files put together)
	if *flagVerbose && len(hashArray) == 1 {
		// We've already printed for the single file. Do nothing.
	} else if *flagBip39 {
		m, err := bip39.NewMnemonic(hash[:])
		if err != nil {
			log.Fatal(err)
		}
		words := strings.Fields(m)
		fmt.Println("  " + strings.Join(words[0:12], " "))
		fmt.Println("  " + strings.Join(words[12:24], " "))
	} else {
		fmt.Printf("  %x\n", hash)
	}
}
