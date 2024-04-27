package main

import "os"

func main() {
	os.Getenv("Exit")
	if true {
		os.Exit(1)
	}
}
