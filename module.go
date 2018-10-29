package main

import (
	"fmt"
	"math/big"
)

type Value interface {
	value()
}

type ValInt struct{ val *big.Int }
type ValFloat struct{ val *big.Rat }
type ValString struct{ val string }
type ValBool struct{ val bool }

func (v ValInt) value()    {}
func (v ValFloat) value()  {}
func (v ValString) value() {}
func (v ValBool) value()   {}

type Scope struct {
	defs map[string]*Token
	vars map[string]Value
}

func NewScope() *Scope {
	return &Scope{
		defs: make(map[string]*Token),
		vars: make(map[string]Value),
	}
}

func (s *Scope) CopyScope() *Scope {
	c := &Scope{
		defs: make(map[string]*Token, len(s.defs)),
		vars: make(map[string]Value, len(s.vars)),
	}
	for k, v := range s.defs {
		c.defs[k] = v
	}
	for k, v := range s.vars {
		c.vars[k] = v
	}
	return c
}

func (s *Scope) Run(stmt Stmt) error {
	switch stmt := stmt.(type) {
	case *StmtVar:
		for _, v := range stmt.Vars {
			if def, exists := s.defs[v.Token.Val]; exists {
				return NewRuntimeError(v.Token,
					"Variable %v already defined on file %#v, line %d",
					v.Token.Val, def.Line.Filename, def.Line.Lineno)
			}
		}
		for _, v := range stmt.Vars {
			s.defs[v.Token.Val] = v.Token
		}
		return nil
	case *StmtAssignment:
		if _, exists := s.defs[stmt.Lhs.Token.Val]; !exists {
			return NewRuntimeError(stmt.Lhs.Token,
				"Variable %v not defined", stmt.Lhs.Token.Val)
		}
		val, err := s.Eval(stmt.Rhs)
		if err != nil {
			return err
		}
		s.vars[stmt.Lhs.Token.Val] = val
		return nil
	case *StmtProcCall:
		panic("TODO")
	case *StmtIf:
		panic("TODO")
	case *StmtWhile:
		panic("TODO")
	case *StmtImport:
		panic("TODO")
	case *StmtUnimport:
		panic("TODO")
	case *StmtUndefine:
		panic("TODO")
	case *StmtExport:
		panic("TODO")
	case *StmtFuncDef:
		panic("TODO")
	case *StmtProcDef:
		panic("TODO")
	case *StmtControl:
		panic("TODO")
	case *StmtReturn:
		panic("TODO")
	default:
		panic(fmt.Sprintf("unsupported statement: %#T", stmt))
	}
	// unreachable
}

func (s *Scope) Eval(expr Expr) (Value, error) {
	switch expr := expr.(type) {
	case *ExprVar:
		name := expr.Var.Token.Val
		if _, defined := s.defs[name]; !defined {
			return nil, NewRuntimeError(expr.Token,
				"Variable %v not defined", name)
		}
		rv := s.vars[name]
		if rv == nil {
			return nil, NewRuntimeError(expr.Token,
				"Variable %v defined but not initialized", name)
		}
		return rv, nil
	case *ExprString:
		return ValString{val: expr.Val}, nil
	case *ExprInt:
		return ValInt{val: expr.Val}, nil
	case *ExprFloat:
		return ValFloat{val: expr.Val}, nil
	case *ExprBool:
		return ValBool{val: expr.Val}, nil
	case *ExprOp:
		panic("TODO")
	case *ExprNot:
		panic("TODO")
	case *ExprIndex:
		panic("TODO")
	case *ExprFuncCall:
		panic("TODO")
	case *ExprNegative:
		panic("TODO")
	default:
		panic(fmt.Sprintf("unsupported expression: %#T", expr))
	}
	// unreachable
}
