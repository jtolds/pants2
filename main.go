package main

import (
	"flag"
	"os"
)

func main() {
	err := Main()
	if err != nil {
		panic(err)
	}
}

func Main() error {
	flag.Parse()
	a := NewApp()
	a.Import("std", StdLib)
	file := flag.Arg(0)
	if file != "" {
		_, err := a.LoadFile(file)
		return err
	}
	_, err := a.LoadInteractive(os.Stdin, os.Stderr)
	return err
}
