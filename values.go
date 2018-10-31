package main

import (
	"fmt"
	"math/big"
	"strings"
)

type Value interface {
	String() string
	value()
}

type ValInt struct{ val *big.Int }

func (v ValInt) String() string { return v.val.String() }

type ValFloat struct{ val *big.Rat }

func (v ValFloat) String() string {
	rv := strings.TrimRight(v.val.FloatString(10), "0")
	if strings.HasSuffix(rv, ".") {
		rv += "0"
	}
	return rv
}

type ValString struct{ val string }

func (v ValString) String() string { return v.val }

type ValBool struct{ val bool }

func (v ValBool) String() string { return fmt.Sprint(v.val) }

type ValProc interface {
	Call(args []Value) error
	Value
}

type ProcCB func([]Value) error

func (f ProcCB) value()                  {}
func (f ProcCB) Call(args []Value) error { return f(args) }
func (f ProcCB) String() string          { return "<builtin>" }

type ValFunc interface {
	Call(args []Value) (Value, error)
	Value
}

func (v ValInt) value()    {}
func (v ValFloat) value()  {}
func (v ValString) value() {}
func (v ValBool) value()   {}

type ValueCell struct {
	Def *Line
	Val Value
}
