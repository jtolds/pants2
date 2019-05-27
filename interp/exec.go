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

func lookupVar(s Scope, v *ast.Var) *ValueCell {
	if v.Depth > 0 {
		return s.LookupDepth(v.Token.Val, v.Depth-1)
	}
	vc, depth := s.Lookup(v.Token.Val)
	v.Depth = depth + 1
	return vc
}

func runVar(s Scope, stmt *ast.StmtVar) error {
	for _, v := range stmt.Vars {
		if d := lookupVar(s, v); d != nil {
			return NewRuntimeError(v.Token,
				"Variable %v already defined on file %#v, line %d",
				v.Token.Val, d.Def.Filename, d.Def.Lineno)
		}
	}
	for _, v := range stmt.Vars {
		var val Value
		if v.Expr != nil {
			var err error
			val, err = Eval(s, v.Expr)
			if err != nil {
				return err
			}
		}
		s.Define(v.Token.Val, &ValueCell{Def: v.Token.Line, Val: val})
	}
	return nil
}

func runAssignment(s Scope, stmt *ast.StmtAssignment) error {
	if d := lookupVar(s, stmt.Lhs); d == nil {
		return NewRuntimeError(stmt.Lhs.Token,
			"Variable %v not defined", stmt.Lhs.Token.Val)
	}
	val, err := Eval(s, stmt.Rhs)
	if err != nil {
		return err
	}
	lookupVar(s, stmt.Lhs).Val = val
	return nil
}

func runProcCall(s Scope, stmt *ast.StmtProcCall) error {
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
}

func runIf(s Scope, stmt *ast.StmtIf) error {
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
			var sf ForkScope
			sf.Init(s)
			return RunAll(&sf, stmt.Body)
		}
	} else {
		if len(stmt.Else) > 0 {
			var sf ForkScope
			sf.Init(s)
			return RunAll(&sf, stmt.Else)
		}
	}
	return nil
}

func runWhile(s Scope, stmt *ast.StmtWhile) error {
	var sf ForkScope
	for {
		sf.Init(s)
		test, err := Eval(&sf, stmt.Test)
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
		err = RunAll(&sf, stmt.Body)
		if err != nil {
			if IsControlErrorType(err, CtrlBreak) {
				return nil
			}
			if !IsControlErrorType(err, CtrlNext) {
				return err
			}
		}
	}
}

func runProcDef(s Scope, stmt *ast.StmtProcDef) error {
	if d := lookupVar(s, stmt.Name); d != nil {
		return NewRuntimeError(stmt.Name.Token,
			"Variable %v already defined on file %#v, line %d",
			stmt.Name.Token.Val, d.Def.Filename, d.Def.Lineno)
	}
	s.Define(stmt.Name.Token.Val, &ValueCell{Def: stmt.Token.Line})
	// the proc flattens the scope, but the scope it flatten needs the
	// proc defined so recursion works
	lookupVar(s, stmt.Name).Val = &UserProc{
		def:   stmt.Token,
		name:  stmt.Name.Token.Val,
		scope: s.Flatten(),
		args:  stmt.Args,
		body:  stmt.Body}
	return nil
}

func runUndefine(s Scope, stmt *ast.StmtUndefine) error {
	for _, v := range stmt.Vars {
		if d := lookupVar(s, v); d == nil {
			return NewRuntimeError(v.Token,
				"Variable %v already not defined", v.Token.Val)
		}
	}
	for _, v := range stmt.Vars {
		s.Remove(v.Token.Val)
	}
	return nil
}

func runFuncDef(s Scope, stmt *ast.StmtFuncDef) error {
	if d := lookupVar(s, stmt.Name); d != nil {
		return NewRuntimeError(stmt.Name.Token,
			"Variable %v already defined on file %#v, line %d",
			stmt.Name.Token.Val, d.Def.Filename, d.Def.Lineno)
	}
	s.Define(stmt.Name.Token.Val, &ValueCell{Def: stmt.Token.Line})
	// the func flattens the scope, but the scope it flatten needs the
	// func defined so recursion works
	lookupVar(s, stmt.Name).Val = &UserFunc{
		def:   stmt.Token,
		name:  stmt.Name.Token.Val,
		scope: s.Flatten(),
		args:  stmt.Args,
		body:  stmt.Body}
	return nil
}

func runReturn(s Scope, stmt *ast.StmtReturn) error {
	val, err := Eval(s, stmt.Val)
	if err != nil {
		return err
	}
	return NewControlError(stmt.Token, val)
}

func runExport(s Scope, stmt *ast.StmtExport) error {
	err := s.Export(stmt)
	if err != nil {
		return NewRuntimeError(stmt.Token, "%s", err.Error())
	}
	return nil
}

func runImport(s Scope, stmt *ast.StmtImport) error {
	var prefix string
	if stmt.Prefix != nil {
		prefix = stmt.Prefix.Token.Val
	}
	err := s.Import(stmt.Path.Val, prefix)
	if err != nil {
		return NewRuntimeError(stmt.Token, "%s", err.Error())
	}
	return nil
}

func runUnimport(s Scope, stmt *ast.StmtUnimport) error {
	err := s.Unimport(stmt.Path.Val)
	if err != nil {
		return NewRuntimeError(stmt.Token, "%s", err.Error())
	}
	return nil
}

func Run(s Scope, stmt ast.Stmt) error {
	switch stmt := stmt.(type) {
	case *ast.StmtVar:
		return runVar(s, stmt)
	case *ast.StmtAssignment:
		return runAssignment(s, stmt)
	case *ast.StmtProcCall:
		return runProcCall(s, stmt)
	case *ast.StmtIf:
		return runIf(s, stmt)
	case *ast.StmtWhile:
		return runWhile(s, stmt)
	case *ast.StmtProcDef:
		return runProcDef(s, stmt)
	case *ast.StmtUndefine:
		return runUndefine(s, stmt)
	case *ast.StmtFuncDef:
		return runFuncDef(s, stmt)
	case *ast.StmtReturn:
		return runReturn(s, stmt)
	case *ast.StmtExport:
		return runExport(s, stmt)
	case *ast.StmtImport:
		return runImport(s, stmt)
	case *ast.StmtUnimport:
		return runUnimport(s, stmt)

	case *ast.StmtControl:
		return NewControlError(stmt.Token, nil)

	default:
		panic(fmt.Sprintf("unsupported statement: %#v", stmt))
	}
	// unreachable
}

func evalVar(s Scope, expr *ast.ExprVar) (Value, error) {
	rv := lookupVar(s, expr.Var)
	name := expr.Var.Token.Val
	if rv == nil {
		return nil, NewRuntimeError(expr.Token,
			"Variable %v not defined", name)
	}
	if rv.Val == nil {
		return nil, NewRuntimeError(expr.Token,
			"Variable %v defined but not initialized", name)
	}
	return rv.Val, nil
}

func evalOp(s Scope, expr *ast.ExprOp) (Value, error) {
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
	if expr.Method == nil {
		var found bool
		expr.Method, found = operations[expr.Op.Type]
		if !found {
			return nil, unsupportedOp(expr.Token, expr.Op.Type, left, right)
		}
	}
	return expr.Method(expr.Token, left, right)
}

func evalFuncCall(s Scope, expr *ast.ExprFuncCall) (Value, error) {
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
}

func evalNot(s Scope, expr *ast.ExprNot) (Value, error) {
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
}

func evalNegative(s Scope, expr *ast.ExprNegative) (Value, error) {
	val, err := Eval(s, expr.Expr)
	if err != nil {
		return nil, err
	}
	switch val := val.(type) {
	case ValNumber:
		rv := ValNumber{}
		rv.Val.Neg(&val.Val)
		return rv, nil
	default:
		return nil, NewRuntimeError(expr.Token,
			"negative requires a number, got %#v instead.", val)
	}
}

func Eval(s Scope, expr ast.Expr) (Value, error) {
	switch expr := expr.(type) {
	case *ast.ExprVar:
		return evalVar(s, expr)
	case *ast.ExprString:
		return ValString{Val: expr.Val}, nil
	case *ast.ExprNumber:
		return ValNumber{Val: expr.Val}, nil
	case *ast.ExprBool:
		return ValBool{Val: expr.Val}, nil
	case *ast.ExprNot:
		return evalNot(s, expr)
	case *ast.ExprNegative:
		return evalNegative(s, expr)
	case *ast.ExprOp:
		return evalOp(s, expr)
	case *ast.ExprIndex:
		panic("TODO")
	case *ast.ExprFuncCall:
		return evalFuncCall(s, expr)
	default:
		panic(fmt.Sprintf("unsupported expression: %#v", expr))
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
