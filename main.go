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
	flagVerbose := flag.Bool("verbose", false, "Show sha256 or mnenomic for every file in folder")
	flagMakeCopy := flag.Bool("makecopy", false, "Create a \\.bip39\\ folder in current directory, and copy read only renamed files there")
	flagO3de := flag.Bool("o3de", false, "Create a SHA256SUMS file suitable for creating an o3de package")
	flagBip39 := flag.Bool("bip39", false, "Shows BIP39 mnenomic instead of sha256")
	flag.Parse()
	flagFilename := ""
	if len(flag.Args()) != 1 {
		log.Fatal("Please provide one filename")
		//flagFilename = "."
	} else {
		flagFilename = flag.Args()[0]
	}

	if *flagO3de && (*flagBip39 || *flagMakeCopy) {
		log.Fatal("-o3de is incompatible with -bip39 or -makecopy")
	}

	if *flagMakeCopy && flagFilename == "." {
		log.Fatal("-makecopy incompatible with folder name \".\"")
	}

	var outputDir string
	if *flagBip39 {
		outputDir = ".bip39"
	} else {
		outputDir = ".sha256"
	}

	var err error
	if *flagMakeCopy {
		// Delete the existing folder .bip39 or .sha256 in current directory
		err := os.RemoveAll(outputDir)
		if err != nil {
			log.Fatal(err)
		}
		// Make a folder .bip39 or .sha256 in current directory
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	var sha256SumsFile *os.File
	if *flagO3de {
		sha256SumsFile, err = os.Create("SHA256SUMS")
		if err != nil {
			log.Fatal(err)
		}
		defer sha256SumsFile.Close()
	}

	// Is flagFilename a file or folder?
	fileInfo, err := os.Stat(flagFilename)
	if err != nil {
		log.Fatal(err)
	}
	if fileInfo.IsDir() {
		originalDir := flagFilename
		var hashArray [][]byte

		// Walk all the files/folders from the root folder named originalDir
		err := filepath.Walk(originalDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err)
				return err
			}

			if info.IsDir() {
				// It's a folder. Reproduce it in \.bip39\ ?
				if *flagMakeCopy {
					err = os.MkdirAll(outputDir+"\\"+path, 0755)
					if err != nil {
						log.Fatal(err)
					}
				}
			} else if info.Name() == "SHA256SUMS" {
				// Skip the SHA256SUMS file
			} else if info.Name() == "shafolder.exe" {
				// Skip the executable of this program in case someone put it there to run it
			} else {
				// It's file we're interested in

				// Get the file's hash
				fileHash := fileSha256(path)

				// Put the sha256 hash of the file into hashArray
				hashArray = append(hashArray, fileHash)

				// Create full hash string (hex sha256 or english bip39)
				// And short mnenomic (6 digit hex sha256 or two english words)
				fullHash, partialHash := fullPartialHash(fileHash, *flagBip39)

				// Print the file path and full hash?
				if *flagVerbose {
					fmt.Println("FILE: ", path)
					fmt.Println(fullHash)
				}

				// Output the full hash and path to SHA256SUMS file?
				if *flagO3de {
					sha256SumsFile.WriteString(fullHash)
					sha256SumsFile.WriteString(" *")
					sha256SumsFile.WriteString(path)
					sha256SumsFile.WriteString("\n")
				}

				// Rename the file and put it somewhere in \.bip39\
				if *flagMakeCopy {
					newFilename := partialHash + " " + info.Name()
					newPath := outputDir + "\\" + filepath.Dir(path) + "\\" + newFilename
					copyFile(newPath, path)
				}
			}
			return nil
		})

		// Did the walk fail?
		if err != nil {
			log.Fatal(err)
		}

		// Now to work out and print the summary hash (the hash of sorted hashes)

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
		var fullHash string
		var partialHash string
		if *flagVerbose && len(hashArray) == 1 {
			// We've already printed for the single file. Do nothing.
		} else {
			fullHash, partialHash = fullPartialHash(hash, *flagBip39)
			fmt.Println(fullHash)
		}

		// Rename the output folder?
		if *flagMakeCopy {
			oldFolder := outputDir + "\\" + originalDir
			newFolder := outputDir + "\\" + partialHash + " " + originalDir
			err = os.Rename(oldFolder, newFolder)
			if err != nil {
				log.Fatal(err)
			}
		}

	} else {
		// It's just one file

		// Get the file's hash
		fileHash := fileSha256(flagFilename)

		// Create full hash string (hex sha256 or english bip39)
		// And short mnenomic (6 digit hex sha256 or two english words)
		fullHash, partialHash := fullPartialHash(fileHash, *flagBip39)

		// Print the file path and full hash?
		if *flagVerbose {
			fmt.Println("FILE: ", flagFilename)
		}
		fmt.Println(fullHash)

		// Rename the file and put it somewhere in \.bip39\
		if *flagMakeCopy {
			newFilename := partialHash + " " + fileInfo.Name()
			newPath := outputDir + "\\" + newFilename
			copyFile(newPath, flagFilename)
		}
	}

	if *flagMakeCopy {
		allContentsReadOnly(outputDir)
	}

	if *flagO3de {
		fmt.Println("-o3de: Hashes and filenames written to SHA256SUMS file suitable for O3DE package")
	}
}

func fileSha256(path string) []byte {
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

	return h.Sum(nil)
}

func fullPartialHash(theHash []byte, bip39flag bool) (full string, partial string) {
	var fullHash string
	var partialHash string
	if bip39flag {
		// Turn the 256 bit sha256 hash into a 24 word mnemonic
		m, err := bip39.NewMnemonic(theHash)
		if err != nil {
			log.Fatal(err)
		}
		words := strings.Fields(m)

		// Return full bip39 as two lines
		fullHash += "  " + strings.Join(words[0:12], " ") + "\n"
		fullHash += "  " + strings.Join(words[12:24], " ")

		// Construct the two word bip39 mnenomic prefix
		partialHash = strcase.ToCamel(words[0] + " " + words[1])
	} else {
		fullHash = fmt.Sprintf("  %x", theHash)
		partialHash = fullHash[2:8]
	}
	return fullHash, partialHash
}

func copyFile(destPath string, srcPath string) {
	source, err := os.Open(srcPath)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()

	destination, err := os.Create(destPath)
	if err != nil {
		log.Fatal(err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		log.Fatal(err)
	}
}

func allContentsReadOnly(folder string) {
	// Walk all the files/folders from the root folder named originalDir
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if path != folder {
			return os.Chmod(path, 0444)
		}
		return nil
	})
	// Was the walk ok?
	if err != nil {
		log.Fatal(err)
	}
}
