package main

import (
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/jtolds/pants2/ast"
	"github.com/jtolds/pants2/interp"
)

func handleErr(err error) {
	if err == nil {
		return
	}
	if interp.IsHandledError(err) {
		_, err = fmt.Println(err)
		if err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}
}

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
	m := interp.NewScope()
	m.Define("print", interp.ProcCB(Print))
	m.Define("time", interp.FuncCB(Time))
	ts := ast.NewTokenSource(ast.NewReaderLineSource("<stdin>", os.Stdin,
		func() error {
			_, err := fmt.Printf("> ")
			return err
		}))
	for {
		stmt, err := ast.ParseStatement(ts)
		if err != nil {
			if err == io.EOF {
				break
			}
			handleErr(err)
			ts.ResetLine()
			continue
		}
		err = m.Run(stmt)
		if err != nil {
			handleErr(err)
			ts.ResetLine()
		}
	}
}
