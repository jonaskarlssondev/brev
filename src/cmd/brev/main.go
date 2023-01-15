package main

import (
	brev "brev/server"
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	s := brev.NewBrevServer()
	return s.ListenAndServe(":80")
}
