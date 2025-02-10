package main

import "os"

func main() {
	otherFunc()
}
func otherFunc() {
	os.Exit(1)
}
