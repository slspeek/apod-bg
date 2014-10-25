package main

import (
	"flag"
	"github.com/slspeek/apod-bg/apod"
	"os"
)

func main() {
	flag.Parse()

	err := apod.Execute()
	if err != nil {
		os.Exit(1)
	}
}
