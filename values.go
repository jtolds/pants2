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

type ValNumber struct{ val *big.Rat }

func (v ValNumber) String() string {
	return strings.TrimRight(strings.TrimRight(v.val.FloatString(10), "0"), ".")
}

type ValString struct{ val string }

func (v ValString) String() string { return v.val }

type ValBool struct{ val bool }

func (v ValBool) String() string { return fmt.Sprint(v.val) }

type ValProc interface {
	Call(t *Token, args []Value) error
	Value
}

type UserProc struct {
	name  string
	scope *Scope
	args  []*Var
	body  []Stmt
}

func (p *UserProc) value()         {}
func (p *UserProc) String() string { return p.name }
func (p *UserProc) Call(t *Token, args []Value) error {
	if len(args) != len(p.args) {
		return NewRuntimeError(t,
			"Expected %d arguments but got %d", len(p.args), len(args))
	}
	for _, arg := range p.args {
		if d, exists := p.scope.vars[arg.Token.Val]; exists {
			return NewRuntimeError(arg.Token,
				"Variable %v already defined on file %#v, line %d",
				arg.Token.Val, d.Def.Filename, d.Def.Lineno)
		}
	}
	s := p.scope.Copy()
	for i := range args {
		s.vars[p.args[i].Token.Val] = &ValueCell{
			Def: p.args[i].Token.Line,
			Val: args[i],
		}
	}
	return s.RunAll(p.body)
}

type ProcCB func([]Value) error

func (f ProcCB) value()                            {}
func (f ProcCB) String() string                    { return "<builtin>" }
func (f ProcCB) Call(t *Token, args []Value) error { return f(args) }

type ValFunc interface {
	Call(t *Token, args []Value) (Value, error)
	Value
}

func (v ValNumber) value() {}
func (v ValString) value() {}
func (v ValBool) value()   {}

type ValueCell struct {
	Def *Line
	Val Value
}
