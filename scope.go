package main

import (
	"fmt"
	"math/big"
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
		return proc.Call(stmt.Token, args)
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
		for {
			test, err := s.Eval(stmt.Test)
			if err != nil {
				return err
			}
			testbool, ok := test.(ValBool)
			if !ok {
				return NewRuntimeError(stmt.Token,
					"while statement requires a truth value, got %#v instead.", test)
			}
			if !testbool.val {
				return nil
			}
			err = s.Copy().RunAll(stmt.Body)
			if err != nil {
				return err
			}
		}
	case *StmtProcDef:
		if d, exists := s.vars[stmt.Name.Token.Val]; exists {
			return NewRuntimeError(stmt.Name.Token,
				"Variable %v already defined on file %#v, line %d",
				stmt.Name.Token.Val, d.Def.Filename, d.Def.Lineno)
		}
		s.vars[stmt.Name.Token.Val] = &ValueCell{
			Def: stmt.Token.Line,
			Val: &UserProc{
				name:  stmt.Name.Token.Val,
				scope: s,
				args:  stmt.Args,
				body:  stmt.Body}}
		return nil
	case *StmtUndefine:
		for _, v := range stmt.Vars {
			if _, exists := s.vars[v.Token.Val]; !exists {
				return NewRuntimeError(v.Token,
					"Variable %v already not defined", v.Token.Val)
			}
		}
		for _, v := range stmt.Vars {
			delete(s.vars, v.Token.Val)
		}
		return nil
	case *StmtImport:
		panic("TODO")
	case *StmtUnimport:
		panic("TODO")
	case *StmtExport:
		panic("TODO")
	case *StmtFuncDef:
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
	case *ExprNumber:
		return ValNumber{val: expr.Val}, nil
	case *ExprBool:
		return ValBool{val: expr.Val}, nil
	case *ExprNot:
		test, err := s.Eval(expr.Expr)
		if err != nil {
			return nil, err
		}
		testbool, ok := test.(ValBool)
		if !ok {
			return nil, NewRuntimeError(expr.Token,
				"not statement requires a truth value, got %#v instead.", test)
		}
		return ValBool{val: !testbool.val}, nil
	case *ExprNegative:
		val, err := s.Eval(expr.Expr)
		if err != nil {
			return nil, err
		}
		switch val := val.(type) {
		case ValNumber:
			n := new(big.Rat)
			return ValNumber{val: n.Neg(val.val)}, nil
		default:
			return nil, NewRuntimeError(expr.Token,
				"negative requires a number, got %#v instead.", val)
		}
	case *ExprOp:
		if expr.Op.Type == "and" || expr.Op.Type == "or" {
			return s.combineBool(expr)
		}
		left, err := s.Eval(expr.Left)
		if err != nil {
			return nil, err
		}
		right, err := s.Eval(expr.Right)
		if err != nil {
			return nil, err
		}
		if expr.Op.Type == "==" || expr.Op.Type == "!=" {
			val := s.equalityTest(expr, left, right)
			if expr.Op.Type == "!=" {
				val = !val
			}
			return ValBool{val: val}, nil
		}
		op, found := operations[opkey{
			op:    expr.Op.Type,
			left:  typename(left),
			right: typename(right),
		}]
		if !found {
			return nil, NewRuntimeError(expr.Token,
				"unsupported operation: %s %s %s",
				typename(left), expr.Op.Type, typename(right))
		}
		return op(expr.Token, left, right)
	case *ExprIndex:
		panic("TODO")
	case *ExprFuncCall:
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

func (s *Scope) equalityTest(expr Expr, left, right Value) bool {
	if typename(left) != typename(right) {
		return false
	}
	switch left.(type) {
	case ValNumber:
		return left.(ValNumber).val.Cmp(right.(ValNumber).val) == 0
	case ValString:
		return left.(ValString).val == right.(ValString).val
	case ValBool:
		return left.(ValBool).val == right.(ValBool).val
	default:
		return false // TODO: throw an error about comparing funcs or procs?
	}
}

type opkey struct{ op, left, right string }

func typename(val interface{}) string { return fmt.Sprintf("%T", val) }

var operations = map[opkey]func(t *Token, left, right Value) (Value, error){
	opkey{"+", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValNumber{
			val: new(big.Rat).Add(left.(ValNumber).val, right.(ValNumber).val)}, nil
	},
	opkey{"-", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValNumber{
			val: new(big.Rat).Sub(left.(ValNumber).val, right.(ValNumber).val)}, nil
	},
	opkey{"*", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValNumber{
			val: new(big.Rat).Mul(left.(ValNumber).val, right.(ValNumber).val)}, nil
	},
	opkey{"/", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *Token, left, right Value) (Value, error) {
		if new(big.Rat).Cmp(right.(ValNumber).val) == 0 {
			return nil, NewRuntimeError(t, "Division by zero")
		}
		return ValNumber{
			val: new(big.Rat).Mul(left.(ValNumber).val,
				new(big.Rat).Inv(right.(ValNumber).val))}, nil
	},
	opkey{"%", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *Token, left, right Value) (Value, error) {
		if new(big.Rat).Cmp(right.(ValNumber).val) == 0 {
			return nil, NewRuntimeError(t, "Division by zero")
		}
		leftdenom := left.(ValNumber).val.Denom()
		rightdenom := right.(ValNumber).val.Denom()
		if !leftdenom.IsInt64() || leftdenom.Int64() != 1 ||
			!rightdenom.IsInt64() || rightdenom.Int64() != 1 {
			return nil, NewRuntimeError(t, "Modulo only works on integers")
		}
		return ValNumber{
			val: new(big.Rat).SetInt(
				new(big.Int).Mod(
					left.(ValNumber).val.Num(),
					right.(ValNumber).val.Num()))}, nil
	},
	opkey{"+", typename(ValString{}), typename(ValString{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValString{val: left.(ValString).val + right.(ValString).val}, nil
	},
	opkey{"<", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValBool{val: left.(ValNumber).val.Cmp(right.(ValNumber).val) < 0}, nil
	},
	opkey{"<=", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValBool{val: left.(ValNumber).val.Cmp(right.(ValNumber).val) <= 0}, nil
	},
	opkey{">", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValBool{val: left.(ValNumber).val.Cmp(right.(ValNumber).val) > 0}, nil
	},
	opkey{">=", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValBool{val: left.(ValNumber).val.Cmp(right.(ValNumber).val) >= 0}, nil
	},
	opkey{"<", typename(ValString{}), typename(ValString{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValBool{val: left.(ValString).val < right.(ValString).val}, nil
	},
	opkey{"<=", typename(ValString{}), typename(ValString{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValBool{val: left.(ValString).val <= right.(ValString).val}, nil
	},
	opkey{">", typename(ValString{}), typename(ValString{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValBool{val: left.(ValString).val >= right.(ValString).val}, nil
	},
	opkey{">=", typename(ValString{}), typename(ValString{})}: func(
		t *Token, left, right Value) (Value, error) {
		return ValBool{val: left.(ValString).val >= right.(ValString).val}, nil
	},
}
