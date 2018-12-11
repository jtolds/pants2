package interp

import (
	"fmt"

	"github.com/jtolds/pants2/ast"
)

func RunAll(s Scope, stmts []ast.Stmt) error {
	for _, stmt := range stmts {
		err := Run(s, stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func Run(s Scope, stmt ast.Stmt) error {
	switch stmt := stmt.(type) {
	case *ast.StmtVar:
		for _, v := range stmt.Vars {
			if d := s.Lookup(v.Token.Val); d != nil {
				return NewRuntimeError(v.Token,
					"Variable %v already defined on file %#v, line %d",
					v.Token.Val, d.Def.Filename, d.Def.Lineno)
			}
		}
		for _, v := range stmt.Vars {
			s.Define(v.Token.Val, &ValueCell{Def: v.Token.Line})
		}
		return nil
	case *ast.StmtAssignment:
		if d := s.Lookup(stmt.Lhs.Token.Val); d == nil {
			return NewRuntimeError(stmt.Lhs.Token,
				"Variable %v not defined", stmt.Lhs.Token.Val)
		}
		val, err := Eval(s, stmt.Rhs)
		if err != nil {
			return err
		}
		s.Lookup(stmt.Lhs.Token.Val).Val = val
		return nil
	case *ast.StmtProcCall:
		procval, err := Eval(s, stmt.Proc)
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
			val, err := Eval(s, arg)
			if err != nil {
				return err
			}
			args = append(args, val)
		}
		return proc.Call(stmt.Token, args)
	case *ast.StmtIf:
		test, err := Eval(s, stmt.Test)
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
				return RunAll(s.Fork(), stmt.Body)
			}
		} else {
			if len(stmt.Else) > 0 {
				return RunAll(s.Fork(), stmt.Else)
			}
		}
		return nil
	case *ast.StmtWhile:
		for {
			sf := s.Fork()
			test, err := Eval(sf, stmt.Test)
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
			err = RunAll(sf, stmt.Body)
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
		if d := s.Lookup(stmt.Name.Token.Val); d != nil {
			return NewRuntimeError(stmt.Name.Token,
				"Variable %v already defined on file %#v, line %d",
				stmt.Name.Token.Val, d.Def.Filename, d.Def.Lineno)
		}
		s.Define(stmt.Name.Token.Val, &ValueCell{Def: stmt.Token.Line})
		// the proc flattens the scope, but the scope it flatten needs the
		// proc defined so recursion works
		s.Lookup(stmt.Name.Token.Val).Val = &UserProc{
			def:   stmt.Token,
			name:  stmt.Name.Token.Val,
			scope: s.Flatten(),
			args:  stmt.Args,
			body:  stmt.Body}
		return nil
	case *ast.StmtUndefine:
		for _, v := range stmt.Vars {
			if d := s.Lookup(v.Token.Val); d == nil {
				return NewRuntimeError(v.Token,
					"Variable %v already not defined", v.Token.Val)
			}
		}
		for _, v := range stmt.Vars {
			s.Remove(v.Token.Val)
		}
		return nil
	case *ast.StmtFuncDef:
		if d := s.Lookup(stmt.Name.Token.Val); d != nil {
			return NewRuntimeError(stmt.Name.Token,
				"Variable %v already defined on file %#v, line %d",
				stmt.Name.Token.Val, d.Def.Filename, d.Def.Lineno)
		}
		s.Define(stmt.Name.Token.Val, &ValueCell{Def: stmt.Token.Line})
		// the func flattens the scope, but the scope it flatten needs the
		// func defined so recursion works
		s.Lookup(stmt.Name.Token.Val).Val = &UserFunc{
			def:   stmt.Token,
			name:  stmt.Name.Token.Val,
			scope: s.Flatten(),
			args:  stmt.Args,
			body:  stmt.Body}
		return nil
	case *ast.StmtControl:
		return NewControlError(stmt.Token, nil)
	case *ast.StmtReturn:
		val, err := Eval(s, stmt.Val)
		if err != nil {
			return err
		}
		return NewControlError(stmt.Token, val)
	case *ast.StmtExport:
		err := s.Export(stmt)
		if err != nil {
			return NewRuntimeError(stmt.Token, "%s", err.Error())
		}
		return nil
	case *ast.StmtImport:
		var prefix string
		if stmt.Prefix != nil {
			prefix = stmt.Prefix.Token.Val
		}
		err := s.Import(stmt.Path.Val, prefix)
		if err != nil {
			return NewRuntimeError(stmt.Token, "%s", err.Error())
		}
		return nil
	case *ast.StmtUnimport:
		err := s.Unimport(stmt.Path.Val)
		if err != nil {
			return NewRuntimeError(stmt.Token, "%s", err.Error())
		}
		return nil
	default:
		panic(fmt.Sprintf("unsupported statement: %#T", stmt))
	}
	// unreachable
}

func Eval(s Scope, expr ast.Expr) (Value, error) {
	switch expr := expr.(type) {
	case *ast.ExprVar:
		name := expr.Var.Token.Val
		rv := s.Lookup(name)
		if rv == nil {
			return nil, NewRuntimeError(expr.Token,
				"Variable %v not defined", name)
		}
		if rv.Val == nil {
			return nil, NewRuntimeError(expr.Token,
				"Variable %v defined but not initialized", name)
		}
		return rv.Val, nil
	case *ast.ExprString:
		return ValString{Val: expr.Val}, nil
	case *ast.ExprNumber:
		return &ValNumber{Val: expr.Val}, nil
	case *ast.ExprBool:
		return ValBool{Val: expr.Val}, nil
	case *ast.ExprNot:
		test, err := Eval(s, expr.Expr)
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
		val, err := Eval(s, expr.Expr)
		if err != nil {
			return nil, err
		}
		switch val := val.(type) {
		case *ValNumber:
			rv := ValNumber{}
			rv.Val.Neg(&val.Val)
			return &rv, nil
		default:
			return nil, NewRuntimeError(expr.Token,
				"negative requires a number, got %#v instead.", val)
		}
	case *ast.ExprOp:
		if expr.Op.Type == "and" || expr.Op.Type == "or" {
			return combineBool(s, expr)
		}
		left, err := Eval(s, expr.Left)
		if err != nil {
			return nil, err
		}
		right, err := Eval(s, expr.Right)
		if err != nil {
			return nil, err
		}
		if expr.Op.Type == "==" || expr.Op.Type == "!=" {
			val := equalityTest(expr, left, right)
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
		funcval, err := Eval(s, expr.Func)
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
			val, err := Eval(s, arg)
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

func combineBool(s Scope, op *ast.ExprOp) (Value, error) {
	left, err := Eval(s, op.Left)
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
	return Eval(s, op.Right)
}
