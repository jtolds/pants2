package main

type Stmt interface {
	statement()
}

type Expr interface {
	expression()
}

type Var struct {
	Token *Token
}

type StmtIf struct {
	Token *Token
	Test  Expr
	Body  []Stmt
}

type StmtElse struct {
	Token *Token
	Body  []Stmt
}

type StmtVar struct {
	Token *Token
	Vars  []*Var
}

type StmtAssignment struct {
	Token *Token
	Lhs   *Var
	Rhs   Expr
}

type StmtWhile struct {
	Token *Token
	Test  Expr
	Body  []Stmt
}

type StmtImport struct {
	Token  *Token
	Path   *ExprString
	Prefix *Var
}

type StmtUnimport struct {
	Token *Token
	Path  *ExprString
}

type StmtUndefine struct {
	Token *Token
	Vars  []*Var
}

type StmtExport struct {
	Token *Token
	Vars  []*Var
}

type StmtFuncDef struct {
	Token *Token
	Name  *Var
	Args  []*Var
	Body  []Stmt
}

type StmtProcDef struct {
	Token *Token
	Name  *Var
	Args  []*Var
	Body  []Stmt
}

type StmtProcCall struct {
	Token *Token
	Proc  Expr
	Args  []Expr
}

type StmtControl struct {
	Token *Token
}

type StmtReturn struct {
	Token *Token
	Val   Expr
}

func (*StmtIf) statement()         {}
func (*StmtElse) statement()       {}
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

type ExprString struct {
	Token *Token
	Val   string
}

type ExprInt struct {
	Token *Token
	Val   int64
}

type ExprFloat struct {
	Token *Token
	Val   float64
}

type ExprBool struct {
	Token *Token
	Val   bool
}

type ExprOp struct {
	Token *Token
	Left  Expr
	Op    *Token
	Right Expr
}

type ExprNot struct {
	Token *Token
	Expr  Expr
}

type ExprIndex struct {
	Token  *Token
	Object Expr
	Index  Expr
}

func (*ExprVar) expression()    {}
func (*ExprString) expression() {}
func (*ExprInt) expression()    {}
func (*ExprFloat) expression()  {}
func (*ExprBool) expression()   {}
func (*ExprOp) expression()     {}
func (*ExprNot) expression()    {}
func (*ExprIndex) expression()  {}
