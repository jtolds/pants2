package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/jtolds/pants2/interp"
)

func Time(args []interp.Value) (interp.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("unexpected arguments")
	}
	return interp.ValNumber{Val: big.NewRat(time.Now().UnixNano(), 1)}, nil
}

func Input(args []interp.Value) (interp.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("unexpected arguments")
	}
	var b [1]byte
	var rv []byte
	for {
		n, err := os.Stdin.Read(b[:])
		if n > 0 {
			rv = append(rv, b[0])
			if b[0] == '\n' {
				break
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return interp.ValString{Val: string(bytes.TrimSpace(rv))}, nil
}

func Print(args []interp.Value) error {
	for i, arg := range args {
		if i > 0 {
			_, err := fmt.Print(" ")
			if err != nil {
				return err
			}
		}
		_, err := fmt.Print(arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func Println(args []interp.Value) error {
	err := Print(args)
	if err != nil {
		return err
	}
	_, err = fmt.Println()
	return err
}

func Number(args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected only one argument")
	}
	switch arg := args[0].(type) {
	case interp.ValString:
		rv, ok := new(big.Rat).SetString(strings.TrimSpace(arg.Val))
		if !ok {
			return nil, fmt.Errorf("could not convert value to number: %#v", arg)
		}
		return interp.ValNumber{Val: rv}, nil
	case interp.ValNumber:
		return arg, nil
	default:
		return nil, fmt.Errorf("could not convert value to number: %#v", arg)
	}
}

func Random(args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected two arguments")
	}
	low, ok := args[0].(interp.ValNumber)
	if !ok {
		return nil, fmt.Errorf("first argument should be a number")
	}
	if big.NewInt(1).Cmp(low.Val.Denom()) != 0 {
		return nil, fmt.Errorf("first argument should be an integer")
	}
	high, ok := args[1].(interp.ValNumber)
	if !ok {
		return nil, fmt.Errorf("second argument should be a number")
	}
	if big.NewInt(1).Cmp(high.Val.Denom()) != 0 {
		return nil, fmt.Errorf("second argument should be an integer")
	}
	z, err := rand.Int(rand.Reader, new(big.Int).Sub(high.Val.Num(), low.Val.Num()))
	if err != nil {
		return nil, err
	}
	return interp.ValNumber{
		Val: new(big.Rat).Add(new(big.Rat).SetInt(z), low.Val)}, nil
}

var (
	StdLib = map[string]interp.Value{
		"print":   interp.ProcCB(Print),
		"println": interp.ProcCB(Println),
		"time":    interp.FuncCB(Time),
		"input":   interp.FuncCB(Input),
		"number":  interp.FuncCB(Number),
		"random":  interp.FuncCB(Random),
		"call":    interp.ProcCB(func([]interp.Value) error { return nil }),
		"CALL":    interp.ProcCB(func([]interp.Value) error { return nil }),
	}
)
