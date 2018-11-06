package interp

import (
	"fmt"
	"math/big"

	"github.com/jtolds/pants2/ast"
)

type ModuleImporterFunc func(path string) (map[string]*ValueCell, error)

func (f ModuleImporterFunc) Import(path string) (map[string]*ValueCell, error) {
	return f(path)
}

type ModuleImporter interface {
	Import(path string) (map[string]*ValueCell, error)
}

type Scope struct {
	vars      map[string]*ValueCell
	exports   map[string]*ValueCell
	importer  ModuleImporter
	unimports map[string]map[string]bool
}

func NewScope(importer ModuleImporter) *Scope {
	return &Scope{
		vars:      map[string]*ValueCell{},
		importer:  importer,
		unimports: map[string]map[string]bool{},
	}
}

func (s *Scope) Define(name string, val Value) {
	s.vars[name] = &ValueCell{
		Def: &ast.Line{Filename: "<builtin>"},
		Val: val,
	}
}

func (s *Scope) EnableExports() {
	if s.exports == nil {
		s.exports = map[string]*ValueCell{}
	}
}

func (s *Scope) Copy() *Scope {
	c := &Scope{
		vars: make(map[string]*ValueCell, len(s.vars)),
		// deliberately don't copy exports
		importer:  s.importer,
		unimports: make(map[string]map[string]bool, len(s.unimports)),
	}
	for k, v := range s.vars {
		c.vars[k] = v
	}
	for mod, vars := range s.unimports {
		c.unimports[mod] = make(map[string]bool, len(vars))
		for k, v := range vars {
			c.unimports[mod][k] = v
		}
	}
	return c
}

func (s *Scope) RunAll(stmts []ast.Stmt) error {
	for _, stmt := range stmts {
		err := s.Run(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scope) Run(stmt ast.Stmt) error {
	switch stmt := stmt.(type) {
	case *ast.StmtVar:
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
	case *ast.StmtAssignment:
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
	case *ast.StmtProcCall:
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
	case *ast.StmtIf:
		test, err := s.Eval(stmt.Test)
		if err != nil {
			return err
		}
		testbool, ok := test.(ValBool)
		if !ok {
			return NewRuntimeError(stmt.Token,
				"if statement requires a truth value, got %#v instead.", test)
		}
		if testbool.Val {
			if len(stmt.Body) > 0 {
				return s.Copy().RunAll(stmt.Body)
			}
		} else {
			if len(stmt.Else) > 0 {
				return s.Copy().RunAll(stmt.Else)
			}
		}
		return nil
	case *ast.StmtWhile:
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
			if !testbool.Val {
				return nil
			}
			err = s.Copy().RunAll(stmt.Body)
			if err != nil {
				if IsControlErrorType(err, CtrlBreak) {
					return nil
				}
				if !IsControlErrorType(err, CtrlNext) {
					return err
				}
			}
		}
	case *ast.StmtProcDef:
		if d, exists := s.vars[stmt.Name.Token.Val]; exists {
			return NewRuntimeError(stmt.Name.Token,
				"Variable %v already defined on file %#v, line %d",
				stmt.Name.Token.Val, d.Def.Filename, d.Def.Lineno)
		}
		s.vars[stmt.Name.Token.Val] = &ValueCell{
			Def: stmt.Token.Line,
			Val: &UserProc{
				def:   stmt.Token,
				name:  stmt.Name.Token.Val,
				scope: s,
				args:  stmt.Args,
				body:  stmt.Body}}
		return nil
	case *ast.StmtUndefine:
		for _, v := range stmt.Vars {
			if _, exists := s.vars[v.Token.Val]; !exists {
				return NewRuntimeError(v.Token,
					"Variable %v already not defined", v.Token.Val)
			}
		}
		for _, v := range stmt.Vars {
			delete(s.vars, v.Token.Val)
			for mod := range s.unimports {
				delete(s.unimports[mod], v.Token.Val)
			}
		}
		return nil
	case *ast.StmtFuncDef:
		if d, exists := s.vars[stmt.Name.Token.Val]; exists {
			return NewRuntimeError(stmt.Name.Token,
				"Variable %v already defined on file %#v, line %d",
				stmt.Name.Token.Val, d.Def.Filename, d.Def.Lineno)
		}
		s.vars[stmt.Name.Token.Val] = &ValueCell{
			Def: stmt.Token.Line,
			Val: &UserFunc{
				def:   stmt.Token,
				name:  stmt.Name.Token.Val,
				scope: s,
				args:  stmt.Args,
				body:  stmt.Body}}
		return nil
	case *ast.StmtControl:
		return NewControlError(stmt.Token, nil)
	case *ast.StmtReturn:
		val, err := s.Eval(stmt.Val)
		if err != nil {
			return err
		}
		return NewControlError(stmt.Token, val)
	case *ast.StmtExport:
		if s.exports == nil {
			return NewRuntimeError(stmt.Token, "Unexpected export")
		}
		for _, v := range stmt.Vars {
			if d, exists := s.exports[v.Token.Val]; exists {
				return NewRuntimeError(v.Token,
					"Exported variable \"%s\" already exported on file %#v, line %d",
					v.Token.Val, d.Def.Filename, d.Def.Lineno)
			}
		}
		for _, v := range stmt.Vars {
			cell, ok := s.vars[v.Token.Val]
			if !ok {
				return NewRuntimeError(v.Token,
					"Variable %v not defined", v.Token.Val)
			}
			s.exports[v.Token.Val] = cell
		}
		return nil
	case *ast.StmtImport:
		exports, err := s.importer.Import(stmt.Path.Val)
		if err != nil {
			return err
		}
		var prefix string
		if stmt.Prefix != nil {
			prefix = stmt.Prefix.Token.Val
		}
		for v := range exports {
			if d, exists := s.vars[prefix+v]; exists {
				return NewRuntimeError(stmt.Token,
					"Export defines %#v, but %#v already defined on file %#v, line %d",
					prefix+v, prefix+v, d.Def.Filename, d.Def.Lineno)
			}
		}
		unimports := make(map[string]bool, len(exports))
		for v, cell := range exports {
			s.vars[prefix+v] = cell
			unimports[prefix+v] = true
		}
		s.unimports[stmt.Path.Val] = unimports
		return nil
	case *ast.StmtUnimport:
		vars, exists := s.unimports[stmt.Path.Val]
		if !exists {
			return NewRuntimeError(stmt.Token, "Module %#v not imported",
				stmt.Path.Val)
		}
		for v := range vars {
			delete(s.vars, v)
		}
		delete(s.unimports, stmt.Path.Val)
		return nil
	default:
		panic(fmt.Sprintf("unsupported statement: %#T", stmt))
	}
	// unreachable
}

func (s *Scope) Eval(expr ast.Expr) (Value, error) {
	switch expr := expr.(type) {
	case *ast.ExprVar:
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
	case *ast.ExprString:
		return ValString{Val: expr.Val}, nil
	case *ast.ExprNumber:
		return ValNumber{Val: expr.Val}, nil
	case *ast.ExprBool:
		return ValBool{Val: expr.Val}, nil
	case *ast.ExprNot:
		test, err := s.Eval(expr.Expr)
		if err != nil {
			return nil, err
		}
		testbool, ok := test.(ValBool)
		if !ok {
			return nil, NewRuntimeError(expr.Token,
				"not statement requires a truth value, got %#v instead.", test)
		}
		return ValBool{Val: !testbool.Val}, nil
	case *ast.ExprNegative:
		val, err := s.Eval(expr.Expr)
		if err != nil {
			return nil, err
		}
		switch val := val.(type) {
		case ValNumber:
			n := new(big.Rat)
			return ValNumber{Val: n.Neg(val.Val)}, nil
		default:
			return nil, NewRuntimeError(expr.Token,
				"negative requires a number, got %#v instead.", val)
		}
	case *ast.ExprOp:
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
			return ValBool{Val: val}, nil
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
	case *ast.ExprIndex:
		panic("TODO")
	case *ast.ExprFuncCall:
		funcval, err := s.Eval(expr.Func)
		if err != nil {
			return nil, err
		}
		fn, ok := funcval.(ValFunc)
		if !ok {
			return nil, NewRuntimeError(expr.Token,
				"Function call without function value. Unexpected value %s",
				funcval)
		}
		args := make([]Value, 0, len(expr.Args))
		for _, arg := range expr.Args {
			val, err := s.Eval(arg)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		return fn.Call(expr.Token, args)
	default:
		panic(fmt.Sprintf("unsupported expression: %#T", expr))
	}
	// unreachable
}

func (s *Scope) combineBool(op *ast.ExprOp) (Value, error) {
	left, err := s.Eval(op.Left)
	if err != nil {
		return nil, err
	}
	leftbool, leftok := left.(ValBool)
	if !leftok {
		return nil, NewRuntimeError(op.Token,
			"Operation \"%s\" expects truth value on left side.", op.Op.Type)
	}
	if (op.Op.Type == "or" && leftbool.Val) ||
		(op.Op.Type == "and" && !leftbool.Val) {
		return leftbool, nil
	}
	return s.Eval(op.Right)
}

func (s *Scope) equalityTest(expr ast.Expr, left, right Value) bool {
	if typename(left) != typename(right) {
		return false
	}
	switch left.(type) {
	case ValNumber:
		return left.(ValNumber).Val.Cmp(right.(ValNumber).Val) == 0
	case ValString:
		return left.(ValString).Val == right.(ValString).Val
	case ValBool:
		return left.(ValBool).Val == right.(ValBool).Val
	default:
		return false // TODO: throw an error about comparing funcs or procs?
	}
}

type opkey struct{ op, left, right string }

func typename(val interface{}) string { return fmt.Sprintf("%T", val) }

var operations = map[opkey]func(t *ast.Token, left, right Value) (
	Value, error){
	opkey{"+", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValNumber{
			Val: new(big.Rat).Add(left.(ValNumber).Val, right.(ValNumber).Val)}, nil
	},
	opkey{"-", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValNumber{
			Val: new(big.Rat).Sub(left.(ValNumber).Val, right.(ValNumber).Val)}, nil
	},
	opkey{"*", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValNumber{
			Val: new(big.Rat).Mul(left.(ValNumber).Val, right.(ValNumber).Val)}, nil
	},
	opkey{"/", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		if new(big.Rat).Cmp(right.(ValNumber).Val) == 0 {
			return nil, NewRuntimeError(t, "Division by zero")
		}
		return ValNumber{
			Val: new(big.Rat).Mul(left.(ValNumber).Val,
				new(big.Rat).Inv(right.(ValNumber).Val))}, nil
	},
	opkey{"%", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		if new(big.Rat).Cmp(right.(ValNumber).Val) == 0 {
			return nil, NewRuntimeError(t, "Division by zero")
		}
		leftdenom := left.(ValNumber).Val.Denom()
		rightdenom := right.(ValNumber).Val.Denom()
		if !leftdenom.IsInt64() || leftdenom.Int64() != 1 ||
			!rightdenom.IsInt64() || rightdenom.Int64() != 1 {
			return nil, NewRuntimeError(t, "Modulo only works on integers")
		}
		return ValNumber{
			Val: new(big.Rat).SetInt(
				new(big.Int).Mod(
					left.(ValNumber).Val.Num(),
					right.(ValNumber).Val.Num()))}, nil
	},
	opkey{"+", typename(ValString{}), typename(ValString{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValString{Val: left.(ValString).Val + right.(ValString).Val}, nil
	},
	opkey{"<", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValNumber).Val.Cmp(right.(ValNumber).Val) < 0}, nil
	},
	opkey{"<=", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValNumber).Val.Cmp(right.(ValNumber).Val) <= 0}, nil
	},
	opkey{">", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValNumber).Val.Cmp(right.(ValNumber).Val) > 0}, nil
	},
	opkey{">=", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValNumber).Val.Cmp(right.(ValNumber).Val) >= 0}, nil
	},
	opkey{"<", typename(ValString{}), typename(ValString{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValString).Val < right.(ValString).Val}, nil
	},
	opkey{"<=", typename(ValString{}), typename(ValString{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValString).Val <= right.(ValString).Val}, nil
	},
	opkey{">", typename(ValString{}), typename(ValString{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValString).Val >= right.(ValString).Val}, nil
	},
	opkey{">=", typename(ValString{}), typename(ValString{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValString).Val >= right.(ValString).Val}, nil
	},
}
