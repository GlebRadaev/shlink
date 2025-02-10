package main

import "os"

func main() {
	select {
	case <-make(chan bool):
		os.Exit(1) // want "os.Exit called in main func in main package"
	}
}
