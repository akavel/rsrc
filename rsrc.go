package main

import (
	//"debug/pe"
	"fmt"
	"os"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) <= 1 {
		return fmt.Errorf("USAGE: %s FILE.exe.manifest\n" +
			"Generates a.res\n")
	}
	return nil
}
