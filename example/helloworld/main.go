package main

import "C"

//export HelloWorld
func HelloWorld() {
	println("Hello, World!")
}

func main() {
	// -biuldmode=c-archive requires a main package.
}
