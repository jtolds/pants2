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

func main() {
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
		handleErr(Run(stmt))
	}
}

func Run(stmt Stmt) error {
	_, err := fmt.Printf("%#v\n", stmt)
	return err
}
