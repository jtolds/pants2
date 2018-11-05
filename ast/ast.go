package ast

import (
	"fmt"
	"math/big"
	"strings"
)

type Stmt interface {
	statement()
	String() string
}

type Expr interface {
	expression()
	String() string
}

type Var struct {
	Token *Token
}

func (v *Var) String() string { return v.Token.Val }

type StmtIf struct {
	Token *Token
	Test  Expr
	Body  []Stmt
	Else  []Stmt
}

func (s *StmtIf) String() string {
	parts := make([]string, 0, len(s.Body)+len(s.Else)+3)
	parts = append(parts, fmt.Sprintf("if %s {\n", s.Test))
	for _, stmt := range s.Body {
		parts = append(parts, stmt.String())
	}
	if len(s.Else) > 0 {
		parts = append(parts, "} else {\n")
		for _, stmt := range s.Else {
			parts = append(parts, stmt.String())
		}
	}
	parts = append(parts, "}\n")
	return strings.Join(parts, "")
}

type StmtVar struct {
	Token *Token
	Vars  []*Var
}

func (s *StmtVar) String() string {
	vars := make([]string, 0, len(s.Vars))
	for _, v := range s.Vars {
		vars = append(vars, v.String())
	}
	return fmt.Sprintf("var %s\n", strings.Join(vars, ", "))
}

type StmtAssignment struct {
	Token *Token
	Lhs   *Var
	Rhs   Expr
}

func (s *StmtAssignment) String() string {
	return fmt.Sprintf("%s = %s\n", s.Lhs, s.Rhs)
}

type StmtWhile struct {
	Token *Token
	Test  Expr
	Body  []Stmt
}

func (s *StmtWhile) String() string {
	parts := make([]string, 0, len(s.Body)+2)
	parts = append(parts, fmt.Sprintf("while %s {\n", s.Test))
	for _, stmt := range s.Body {
		parts = append(parts, stmt.String())
	}
	parts = append(parts, "}\n")
	return strings.Join(parts, "")
}

type StmtImport struct {
	Token  *Token
	Path   *ExprString
	Prefix *Var
}

func (s *StmtImport) String() string {
	if s.Prefix == nil {
		return fmt.Sprintf("import %s\n", s.Path.String())
	}
	return fmt.Sprintf("import %s withprefix %s\n",
		s.Path.String(), s.Prefix.String())
}

type StmtUnimport struct {
	Token *Token
	Path  *ExprString
}

func (s *StmtUnimport) String() string {
	return fmt.Sprintf("unimport %s\n", s.Path.String())
}

type StmtUndefine struct {
	Token *Token
	Vars  []*Var
}

func (s *StmtUndefine) String() string {
	vars := make([]string, 0, len(s.Vars))
	for _, v := range s.Vars {
		vars = append(vars, v.String())
	}
	return fmt.Sprintf("undefine %s\n", strings.Join(vars, ", "))
}

type StmtExport struct {
	Token *Token
	Vars  []*Var
}

func (s *StmtExport) String() string {
	vars := make([]string, 0, len(s.Vars))
	for _, v := range s.Vars {
		vars = append(vars, v.String())
	}
	return fmt.Sprintf("export %s\n", strings.Join(vars, ", "))
}

type StmtFuncDef struct {
	Token *Token
	Name  *Var
	Args  []*Var
	Body  []Stmt
}

func (s *StmtFuncDef) String() string {
	rv := make([]string, 0, len(s.Body)+2)
	args := make([]string, 0, len(s.Args))
	for _, arg := range s.Args {
		args = append(args, arg.Token.Val)
	}
	rv = append(rv, fmt.Sprintf("func %s(%s) {\n",
		s.Name.Token.Val, strings.Join(args, ", ")))
	for _, stmt := range s.Body {
		rv = append(rv, stmt.String())
	}
	rv = append(rv, "}\n")
	return strings.Join(rv, "")
}

type StmtProcDef struct {
	Token *Token
	Name  *Var
	Args  []*Var
	Body  []Stmt
}

func (s *StmtProcDef) String() string {
	rv := make([]string, 0, len(s.Body)+2)
	args := make([]string, 0, len(s.Args))
	for _, arg := range s.Args {
		args = append(args, " "+arg.Token.Val)
	}
	rv = append(rv, fmt.Sprintf("proc %s%s {\n",
		s.Name.Token.Val, strings.Join(args, ",")))
	for _, stmt := range s.Body {
		rv = append(rv, stmt.String())
	}
	rv = append(rv, "}\n")
	return strings.Join(rv, "")
}

type StmtProcCall struct {
	Token *Token
	Proc  Expr
	Args  []Expr
}

func (s *StmtProcCall) String() string {
	rv := make([]string, 0, len(s.Args)+2)
	rv = append(rv, s.Proc.String())
	for i := 0; i < len(s.Args); i++ {
		if i+1 == len(s.Args) {
			rv = append(rv, fmt.Sprintf(" %s", s.Args[i]))
		} else {
			rv = append(rv, fmt.Sprintf(" %s,", s.Args[i]))
		}
	}
	rv = append(rv, "\n")
	return strings.Join(rv, "")
}

type StmtControl struct {
	Token *Token
}

func (s *StmtControl) String() string {
	return s.Token.Val + "\n"
}

type StmtReturn struct {
	Token *Token
	Val   Expr
}

func (s *StmtReturn) String() string {
	return fmt.Sprintf("return %s\n", s.Val)
}

func (*StmtIf) statement()         {}
func (*StmtVar) statement()        {}
func (*StmtAssignment) statement() {}
func (*StmtWhile) statement()      {}
func (*StmtImport) statement()     {}
func (*StmtUnimport) statement()   {}
func (*StmtUndefine) statement()   {}
func (*StmtExport) statement()     {}
func (*StmtFuncDef) statement()    {}
func (*StmtProcDef) statement()    {}
func (*StmtProcCall) statement()   {}
func (*StmtControl) statement()    {}
func (*StmtReturn) statement()     {}

type ExprVar struct {
	Token *Token
	Var   *Var
}

func (e *ExprVar) String() string {
	return e.Var.String()
}

type ExprString struct {
	Token *Token
	Val   string
}

func (e *ExprString) String() string {
	return fmt.Sprintf("%#v", e.Val)
}

type ExprNumber struct {
	Token *Token
	Val   *big.Rat
}

func (e *ExprNumber) String() string {
	return fmt.Sprintf("%s", e.Val)
}

type ExprBool struct {
	Token *Token
	Val   bool
}

func (e *ExprBool) String() string {
	return fmt.Sprintf("%v", e.Val)
}

type ExprOp struct {
	Token *Token
	Left  Expr
	Op    *Token
	Right Expr
}

func (e *ExprOp) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left, e.Op.Type, e.Right)
}

type ExprNot struct {
	Token *Token
	Expr  Expr
}

func (e *ExprNot) String() string {
	return fmt.Sprintf("not %s", e.Expr)
}

type ExprIndex struct {
	Token  *Token
	Object Expr
	Index  Expr
}

func (e *ExprIndex) String() string {
	return fmt.Sprintf("%s[%s]", e.Object, e.Index)
}

type ExprFuncCall struct {
	Token *Token
	Func  Expr
	Args  []Expr
}

func (e *ExprFuncCall) String() string {
	rv := make([]string, 0, len(e.Args)+3)
	rv = append(rv, e.Func.String())
	rv = append(rv, "(")
	for i := 0; i < len(e.Args); i++ {
		if i+1 == len(e.Args) {
			rv = append(rv, fmt.Sprintf("%s", e.Args[i]))
		} else {
			rv = append(rv, fmt.Sprintf("%s, ", e.Args[i]))
		}
	}
	rv = append(rv, ")")
	return strings.Join(rv, "")
}

type ExprNegative struct {
	Token *Token
	Expr  Expr
}

func (e *ExprNegative) String() string {
	return fmt.Sprintf("-%s", e.Expr)
}

func (*ExprVar) expression()      {}
func (*ExprString) expression()   {}
func (*ExprNumber) expression()   {}
func (*ExprBool) expression()     {}
func (*ExprOp) expression()       {}
func (*ExprNot) expression()      {}
func (*ExprIndex) expression()    {}
func (*ExprFuncCall) expression() {}
func (*ExprNegative) expression() {}
