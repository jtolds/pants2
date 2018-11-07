package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/jtolds/pants2/interp"
)

func Time(args []interp.Value) (interp.Value, error) {
	return interp.ValNumber{Val: big.NewRat(time.Now().UnixNano(), 1)}, nil
}

func Print(args []interp.Value) error {
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
	err := Main()
	if err != nil {
		panic(err)
	}
}

func Main() error {
	a := NewApp()
	a.Define("print", interp.ProcCB(Print))
	a.Define("time", interp.FuncCB(Time))

	flag.Parse()
	file := flag.Arg(0)
	if file != "" {
		_, err := a.LoadFile(file)
		return err
	}
	_, err := a.LoadInteractive(os.Stdin, os.Stderr)
	return err
}
