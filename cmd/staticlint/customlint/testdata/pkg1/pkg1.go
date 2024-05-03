package main

import "os"

func main() {
	os.Getenv("Exit")
	if true {
		os.Exit(1)
	}
	os.Exit(1) // want "direct calls of os.Exit are not allowed"
}
