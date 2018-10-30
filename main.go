package main

import (
	"fmt"
	"io"
	"os"
)

func handleErr(err error) {
	if err == nil {
		return
	}
	if IsHandledError(err) {
		_, err = fmt.Println(err)
		if err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}
}

func Print(args []Value) error {
	for _, arg := range args {
		_, err := fmt.Print(arg)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Println()
	return err
}

func main() {
	m := NewScope()
	m.Define("print", ProcCB(Print))
	ts := NewTokenSource(NewReaderLineSource("<stdin>", os.Stdin, func() error {
		_, err := fmt.Printf("> ")
		return err
	}))
	for {
		stmt, err := ParseStatement(ts)
		if err != nil {
			if err == io.EOF {
				break
			}
			handleErr(err)
			continue
		}
		handleErr(m.Run(stmt))
	}
}
