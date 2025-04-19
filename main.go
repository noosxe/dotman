package main

import "fmt"

func main() {
	flags := parseFlags()

	if flags.Verbose {
		fmt.Println("Verbose mode enabled")
	}

	fmt.Println("Hello, World!")
}
