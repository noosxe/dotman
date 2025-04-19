package main

import "flag"

type Flags struct {
	Verbose bool
}

func parseFlags() *Flags {
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	return &Flags{Verbose: *verbose}
}
