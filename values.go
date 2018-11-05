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
	def   *Token
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
	err := s.RunAll(p.body)
	if ce, ok := err.(*ControlError); ok {
		switch ce.typ {
		case CtrlBreak, CtrlNext, CtrlReturn:
			return NewRuntimeError(ce.token, "Unexpected \"%s\"", string(ce.typ))
		case CtrlDone:
			return nil
		default:
			panic(fmt.Sprintf("unknown control type: %s", string(ce.typ)))
		}
	}
	return err
}

type ProcCB func([]Value) error

func (f ProcCB) value()                            {}
func (f ProcCB) String() string                    { return "<builtin>" }
func (f ProcCB) Call(t *Token, args []Value) error { return f(args) }

type ValFunc interface {
	Call(t *Token, args []Value) (Value, error)
	Value
}

type UserFunc struct {
	def   *Token
	name  string
	scope *Scope
	args  []*Var
	body  []Stmt
}

func (f *UserFunc) value()         {}
func (f *UserFunc) String() string { return f.name + "()" }
func (f *UserFunc) Call(t *Token, args []Value) (Value, error) {
	if len(args) != len(f.args) {
		return nil, NewRuntimeError(t,
			"Expected %d arguments but got %d", len(f.args), len(args))
	}
	for _, arg := range f.args {
		if d, exists := f.scope.vars[arg.Token.Val]; exists {
			return nil, NewRuntimeError(arg.Token,
				"Variable %v already defined on file %#v, line %d",
				arg.Token.Val, d.Def.Filename, d.Def.Lineno)
		}
	}
	s := f.scope.Copy()
	for i := range args {
		s.vars[f.args[i].Token.Val] = &ValueCell{
			Def: f.args[i].Token.Line,
			Val: args[i],
		}
	}
	err := s.RunAll(f.body)
	if err == nil {
		return nil, NewRuntimeError(f.def,
			"Function exited with no return statement")
	}
	if ce, ok := err.(*ControlError); ok {
		switch ce.typ {
		case CtrlBreak, CtrlNext, CtrlDone:
			return nil, NewRuntimeError(ce.token, "Unexpected \"%s\"", string(ce.typ))
		case CtrlReturn:
			return ce.val, nil
		default:
			panic(fmt.Sprintf("unknown control type: %s", string(ce.typ)))
		}
	}
	return nil, err
}

type FuncCB func([]Value) (Value, error)

func (f FuncCB) value()                                     {}
func (f FuncCB) String() string                             { return "<builtin>" }
func (f FuncCB) Call(t *Token, args []Value) (Value, error) { return f(args) }

func (v ValNumber) value() {}
func (v ValString) value() {}
func (v ValBool) value()   {}

type ValueCell struct {
	Def *Line
	Val Value
}
