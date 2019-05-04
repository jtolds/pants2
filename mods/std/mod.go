package std

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	stdbig "math/big"
	"os"
	"strings"
	"time"

	"github.com/jtolds/pants2/interp"
	"github.com/jtolds/pants2/lib/big"
)

func Time(args []interp.Value) (interp.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("unexpected arguments")
	}
	var rv interp.ValNumber
	rv.Val.SetInt64(time.Now().UnixNano())
	return &rv, nil
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
		var rv interp.ValNumber
		_, ok := rv.Val.SetString(strings.TrimSpace(arg.Val))
		if !ok {
			return nil, fmt.Errorf("could not convert value to number: %#v", arg)
		}
		return &rv, nil
	case *interp.ValNumber:
		return arg, nil
	default:
		return nil, fmt.Errorf("could not convert value to number: %#v", arg)
	}
}

var one = big.NewInt(1)

func Random(args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected two arguments")
	}
	low, ok := args[0].(*interp.ValNumber)
	if !ok {
		return nil, fmt.Errorf("first argument should be a number")
	}
	if one.Cmp(low.Val.Denom()) != 0 {
		return nil, fmt.Errorf("first argument should be an integer")
	}
	high, ok := args[1].(*interp.ValNumber)
	if !ok {
		return nil, fmt.Errorf("second argument should be a number")
	}
	if one.Cmp(high.Val.Denom()) != 0 {
		return nil, fmt.Errorf("second argument should be an integer")
	}

	var r, s stdbig.Int
	r.SetBytes(high.Val.Num().Bytes())
	s.SetBytes(low.Val.Num().Bytes())
	// TODO: make sure r.Sub(&r, &s) is not greater than what fits in int64
	z, err := rand.Int(rand.Reader, r.Sub(&r, &s))
	if err != nil {
		return nil, err
	}
	var rv interp.ValNumber
	var im big.Rat
	var num big.Int
	num.SetBytes(z.Bytes())
	im.SetInt(&num)
	rv.Val.Add(&im, &low.Val)
	return &rv, nil
}

func Mod() (map[string]interp.Value, error) {
	return map[string]interp.Value{
		// "print":   interp.ProcCB(Print),
		// "println": interp.ProcCB(Println),
		"log":    interp.ProcCB(Println),
		"time":   interp.FuncCB(Time),
		"input":  interp.FuncCB(Input),
		"number": interp.FuncCB(Number),
		"random": interp.FuncCB(Random),
		"call":   interp.ProcCB(func([]interp.Value) error { return nil }),
		"CALL":   interp.ProcCB(func([]interp.Value) error { return nil }),
	}, nil
}
