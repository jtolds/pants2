package interp

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/jtolds/pants2/ast"
)

type Value interface {
	String() string
	value()
}

type ValNumber struct{ Val big.Rat }

func (v *ValNumber) String() string {
	return strings.TrimRight(strings.TrimRight(v.Val.FloatString(10), "0"), ".")
}

type ValString struct{ Val string }

func (v ValString) String() string { return v.Val }

type ValBool struct{ Val bool }

func (v ValBool) String() string { return fmt.Sprint(v.Val) }

type ValProc interface {
	Call(t *ast.Token, args []Value) error
	Value
}

type UserProc struct {
	def   *ast.Token
	name  string
	scope Scope
	args  []*ast.Var
	body  []ast.Stmt
}

func (p *UserProc) value()         {}
func (p *UserProc) String() string { return p.name }
func (p *UserProc) Call(t *ast.Token, args []Value) error {
	if len(args) != len(p.args) {
		return NewRuntimeError(t,
			"Expected %d arguments but got %d", len(p.args), len(args))
	}
	for _, arg := range p.args {
		if d := p.scope.Lookup(arg.Token.Val); d != nil {
			return NewRuntimeError(arg.Token,
				"Variable %v already defined on file %#v, line %d",
				arg.Token.Val, d.Def.Filename, d.Def.Lineno)
		}
	}
	s := p.scope.Fork()
	for i := range args {
		s.Define(p.args[i].Token.Val, &ValueCell{
			Def: p.args[i].Token.Line,
			Val: args[i],
		})
	}
	err := RunAll(s, p.body)
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

func (f ProcCB) value()                                {}
func (f ProcCB) String() string                        { return "<builtin>" }
func (f ProcCB) Call(t *ast.Token, args []Value) error { return f(args) }

type ValFunc interface {
	Call(t *ast.Token, args []Value) (Value, error)
	Value
}

type UserFunc struct {
	def   *ast.Token
	name  string
	scope Scope
	args  []*ast.Var
	body  []ast.Stmt
}

func (f *UserFunc) value()         {}
func (f *UserFunc) String() string { return f.name + "()" }
func (f *UserFunc) Call(t *ast.Token, args []Value) (Value, error) {
	if len(args) != len(f.args) {
		return nil, NewRuntimeError(t,
			"Expected %d arguments but got %d", len(f.args), len(args))
	}
	for _, arg := range f.args {
		if d := f.scope.Lookup(arg.Token.Val); d != nil {
			return nil, NewRuntimeError(arg.Token,
				"Variable %v already defined on file %#v, line %d",
				arg.Token.Val, d.Def.Filename, d.Def.Lineno)
		}
	}
	s := f.scope.Fork()
	for i := range args {
		s.Define(f.args[i].Token.Val, &ValueCell{
			Def: f.args[i].Token.Line,
			Val: args[i],
		})
	}
	err := RunAll(s, f.body)
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

func (f FuncCB) value()                                         {}
func (f FuncCB) String() string                                 { return "<builtin>" }
func (f FuncCB) Call(t *ast.Token, args []Value) (Value, error) { return f(args) }

func (v *ValNumber) value() {}
func (v ValString) value()  {}
func (v ValBool) value()    {}

type ValueCell struct {
	Def *ast.Line
	Val Value
}
