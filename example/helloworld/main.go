package main

import "C"

//export HelloWorld
func HelloWorld() {
	println("Hello, World!")
}

func main() {
	// -buildmode=c-archive requires a main package.
}
