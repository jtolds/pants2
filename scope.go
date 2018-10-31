package main

import (
	"fmt"
)

type Scope struct {
	vars map[string]*ValueCell
}

func NewScope() *Scope {
	return &Scope{
		vars: make(map[string]*ValueCell),
	}
}

func (s *Scope) Define(name string, val Value) {
	s.vars[name] = &ValueCell{
		Def: &Line{Filename: "<builtin>"},
		Val: val,
	}
}

func (s *Scope) Copy() *Scope {
	c := &Scope{
		vars: make(map[string]*ValueCell, len(s.vars)),
	}
	for k, v := range s.vars {
		c.vars[k] = v
	}
	return c
}

func (s *Scope) RunAll(stmts []Stmt) error {
	for _, stmt := range stmts {
		err := s.Run(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scope) Run(stmt Stmt) error {
	switch stmt := stmt.(type) {
	case *StmtVar:
		for _, v := range stmt.Vars {
			if d, exists := s.vars[v.Token.Val]; exists {
				return NewRuntimeError(v.Token,
					"Variable %v already defined on file %#v, line %d",
					v.Token.Val, d.Def.Filename, d.Def.Lineno)
			}
		}
		for _, v := range stmt.Vars {
			s.vars[v.Token.Val] = &ValueCell{Def: v.Token.Line}
		}
		return nil
	case *StmtAssignment:
		if _, exists := s.vars[stmt.Lhs.Token.Val]; !exists {
			return NewRuntimeError(stmt.Lhs.Token,
				"Variable %v not defined", stmt.Lhs.Token.Val)
		}
		val, err := s.Eval(stmt.Rhs)
		if err != nil {
			return err
		}
		s.vars[stmt.Lhs.Token.Val].Val = val
		return nil
	case *StmtProcCall:
		procval, err := s.Eval(stmt.Proc)
		if err != nil {
			return err
		}
		proc, ok := procval.(ValProc)
		if !ok {
			return NewRuntimeError(stmt.Token,
				"Procedure call without procedure value. Unexpected value %s",
				procval)
		}
		args := make([]Value, 0, len(stmt.Args))
		for _, arg := range stmt.Args {
			val, err := s.Eval(arg)
			if err != nil {
				return err
			}
			args = append(args, val)
		}
		return proc.Call(args)
	case *StmtIf:
		test, err := s.Eval(stmt.Test)
		if err != nil {
			return err
		}
		testbool, ok := test.(ValBool)
		if !ok {
			return NewRuntimeError(stmt.Token,
				"if statement requires a truth value, got %#v instead.", test)
		}
		if testbool.val {
			if len(stmt.Body) > 0 {
				return s.Copy().RunAll(stmt.Body)
			}
		} else {
			if len(stmt.Else) > 0 {
				return s.Copy().RunAll(stmt.Else)
			}
		}
		return nil
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
		if _, defined := s.vars[name]; !defined {
			return nil, NewRuntimeError(expr.Token,
				"Variable %v not defined", name)
		}
		rv := s.vars[name].Val
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
		if expr.Op.Type == "and" || expr.Op.Type == "or" {
			return s.combineBool(expr)
		}
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

func (s *Scope) combineBool(op *ExprOp) (Value, error) {
	left, err := s.Eval(op.Left)
	if err != nil {
		return nil, err
	}
	leftbool, leftok := left.(ValBool)
	if !leftok {
		return nil, NewRuntimeError(op.Token,
			"Operation \"%s\" expects truth value on left side.", op.Op.Type)
	}
	if (op.Op.Type == "or" && leftbool.val) ||
		(op.Op.Type == "and" && !leftbool.val) {
		return leftbool, nil
	}
	return s.Eval(op.Right)
}
