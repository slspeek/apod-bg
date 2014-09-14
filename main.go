package main

import (
	"flag"
	"github.com/slspeek/apod-bg/apod"
)

func main() {
	flag.Parse()
	apod.Execute()
}
