package main

import "os"

func main() {
	os.Exit(1) // want "os.Exit called in main func in main package"
}
