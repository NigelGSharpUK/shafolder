package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Please provide one filename")
	} else {
		fi, err := os.Stat(os.Args[1])
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Filesize of ", os.Args[1], ": ", fi.Size())
		}
	}
}
