package main

import (
	"io"
	"math/big"
)

func ParseStatement(tokens *TokenSource) (stmt Stmt, err error) {
	var token *Token
loop:
	for {
		token, err = tokens.NextToken()
		if err != nil {
			return nil, err
		}
		switch token.Type {
		case "eof":
			return nil, io.EOF
		case "newline", ";":
		default:
			break loop
		}
	}

	stmt, err = func() (stmt Stmt, err error) {
		if token.Type == "keyword" {
			switch token.Val {
			case "if":
				return parseIf(token, tokens)
			case "var":
				return parseVar(token, tokens)
			case "loop":
				return parseLoop(token, tokens)
			case "while":
				return parseWhile(token, tokens)
			case "import":
				return parseImport(token, tokens)
			case "unimport":
				return parseUnimport(token, tokens)
			case "undefine":
				return parseUndefine(token, tokens)
			case "export":
				return parseExport(token, tokens)
			case "func":
				return parseFunc(token, tokens)
			case "proc":
				return parseProc(token, tokens)
			case "break", "next", "done":
				return &StmtControl{Token: token}, nil
			case "return":
				return parseReturn(token, tokens)
			default:
				return nil, NewSyntaxErrorFromToken(token,
					"Unexpected keyword %#v. Expecting statement.", token.Type)
			}
		}

		if token.Type == "variable" {
			nextToken, err := tokens.NextToken()
			if err != nil {
				return nil, err
			}
			if nextToken.Type == "=" {
				return parseAssignment(token, tokens)
			}
			tokens.Push(nextToken)
			tokens.Push(token)
			return parseProcCall(tokens)
		}

		if token.Type == "(" {
			tokens.Push(token)
			return parseProcCall(tokens)
		}

		return nil, NewSyntaxErrorFromToken(token,
			"Unexpected token %#v. Expecting statement.", token.Type)
	}()
	if err != nil {
		return nil, err
	}
	tok, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	if tok.Type == "eof" || tok.Type == "}" {
		tokens.Push(tok)
		return stmt, nil
	}
	if tok.Type != "newline" && tok.Type != ";" {
		return nil, NewSyntaxErrorFromToken(tok,
			"Unexpected token %#v. Expecting semicolon or newline.", tok.Type)
	}
	return stmt, nil
}

func parseExprOrder1(tokens *TokenSource) (Expr, error) {
	tok, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	switch tok.Type {
	case "variable":
		return &ExprVar{Token: tok, Var: &Var{Token: tok}}, nil
	case "string":
		return &ExprString{Token: tok, Val: tok.Val}, nil
	case "number":
		val := new(big.Rat)
		_, ok := val.SetString(tok.Val)
		if !ok {
			panic("failed to parse tokenized number")
		}
		return &ExprNumber{Token: tok, Val: val}, nil
	case "bool":
		return &ExprBool{Token: tok, Val: tok.Val == "true"}, nil
	case "(":
		expr, err := parseExpression(tokens, true)
		if err != nil {
			return nil, err
		}
		end, err := tokens.NextToken()
		if err != nil {
			return nil, err
		}
		if end.Type != ")" {
			return nil, NewSyntaxErrorFromToken(end,
				"Unexpected token %#v. Expecting closing parenthesis.", end.Type)
		}
		return expr, nil
	default:
		return nil, NewSyntaxErrorFromToken(
			tok, "Unexpected token %#v.", tok.Type)
	}
}

func parseExprOrder2(tokens *TokenSource) (Expr, error) {
	val, err := parseExprOrder1(tokens)
	if err != nil {
		return nil, err
	}
	tok, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	switch tok.Type {
	case "[": // index
		idx, err := parseExpression(tokens, true)
		if err != nil {
			return nil, err
		}
		end, err := tokens.NextToken()
		if err != nil {
			return nil, err
		}
		if end.Type != "]" {
			return nil, NewSyntaxErrorFromToken(end,
				"Unexpected token %#v. Expecting closing brace \"]\".", end.Type)
		}
		return &ExprIndex{
			Token:  tok,
			Object: val,
			Index:  idx,
		}, nil
	case "(": // function call
		var args []Expr
		for {
			arg, err := parseExpression(tokens, true)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			end, err := tokens.NextToken()
			if err != nil {
				return nil, err
			}
			if end.Type == ")" {
				break
			}
			if end.Type != "," {
				return nil, NewSyntaxErrorFromToken(end,
					"Unexpected token %#v. Expecting closing parenthesis or comma.",
					end.Type)
			}
		}
		return &ExprFuncCall{
			Token: tok,
			Func:  val,
			Args:  args,
		}, nil
	default:
		tokens.Push(tok)
		return val, err
	}
}

func nextToken(t *TokenSource, ignoreNewlines bool) (rv *Token, err error) {
	for {
		tok, err := t.NextToken()
		if err != nil {
			return nil, err
		}
		if !ignoreNewlines || tok.Type != "newline" {
			return tok, nil
		}
	}
}

func parseExprOrder3(tokens *TokenSource, ignoreNewlines bool) (Expr, error) {
	tok, err := nextToken(tokens, ignoreNewlines)
	if err != nil {
		return nil, err
	}
	// TODO: these are probably in the wrong order of ops spot
	switch tok.Type {
	case "-":
		val, err := parseExprOrder3(tokens, ignoreNewlines)
		if err != nil {
			return nil, err
		}
		return &ExprNegative{Token: tok, Expr: val}, nil
	case "not":
		val, err := parseExprOrder3(tokens, ignoreNewlines)
		if err != nil {
			return nil, err
		}
		return &ExprNot{Token: tok, Expr: val}, nil
	}
	tokens.Push(tok)
	return parseExprOrder2(tokens)
}

func parseExprOrder4(tokens *TokenSource, ignoreNewlines bool,
	ops []map[string]bool) (Expr, error) {
	if len(ops) == 0 {
		return parseExprOrder3(tokens, ignoreNewlines)
	}
	val, err := parseExprOrder4(tokens, ignoreNewlines, ops[1:])
	if err != nil {
		return nil, err
	}
	for {
		opToken, err := nextToken(tokens, ignoreNewlines)
		if err != nil {
			return nil, err
		}
		if ops[0][opToken.Type] {
			next, err := parseExprOrder4(tokens, ignoreNewlines, ops[1:])
			if err != nil {
				return nil, err
			}
			val = &ExprOp{
				Token: opToken,
				Left:  val,
				Op:    opToken,
				Right: next,
			}
		} else {
			tokens.Push(opToken)
			return val, nil
		}
	}
}

func parseExpression(tokens *TokenSource, ignoreNewlines bool) (Expr, error) {
	return parseExprOrder4(tokens, ignoreNewlines,
		[]map[string]bool{
			map[string]bool{
				"and": true, "or": true,
			}, map[string]bool{
				"==": true, "!=": true, "<": true, "<=": true, ">": true, ">=": true,
			}, map[string]bool{
				"+": true, "-": true,
			}, map[string]bool{
				"*": true, "%": true, "/": true,
			},
		})
}

func parseStatementBlock(tokens *TokenSource) (rv []Stmt, err error) {
	leftbrace, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	if leftbrace.Type != "{" {
		return nil, NewSyntaxErrorFromToken(leftbrace,
			"Unexpected token %#v. Expecting '{'", leftbrace.Type)
	}
	for {
		var rightbrace *Token
	loop:
		for {
			rightbrace, err = tokens.NextToken()
			if err != nil {
				return nil, err
			}
			switch rightbrace.Type {
			case "newline", ";":
			default:
				break loop
			}
		}

		if rightbrace.Type == "}" {
			return rv, nil
		}
		tokens.Push(rightbrace)
		stmt, err := ParseStatement(tokens)
		if err != nil {
			return nil, err
		}
		rv = append(rv, stmt)
	}
}

// IF <expression> { <statement>* } [ ELSE { <statement>* } ]
func parseIf(start *Token, tokens *TokenSource) (Stmt, error) {
	expr, err := parseExpression(tokens, false)
	if err != nil {
		return nil, err
	}
	stmts, err := parseStatementBlock(tokens)
	if err != nil {
		return nil, err
	}
	tok, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	if tok.Type != "keyword" || tok.Val != "else" {
		tokens.Push(tok)
		return &StmtIf{
			Token: start,
			Test:  expr,
			Body:  stmts,
		}, nil
	}
	elseBody, err := parseStatementBlock(tokens)
	if err != nil {
		return nil, err
	}
	return &StmtIf{
		Token: start,
		Test:  expr,
		Body:  stmts,
		Else:  elseBody,
	}, nil
}

func parseVarList(tokens *TokenSource, ignoreNewlines bool) ([]*Var, error) {
	v, err := nextToken(tokens, ignoreNewlines)
	if err != nil {
		return nil, err
	}
	if v.Type != "variable" {
		return nil, NewSyntaxErrorFromToken(
			v, "Unexpected token %#v. Expecting variable", v.Type)
	}
	vars := []*Var{&Var{Token: v}}
	for {
		comma, err := nextToken(tokens, ignoreNewlines)
		if err != nil {
			return nil, err
		}
		if comma.Type != "," {
			tokens.Push(comma)
			break
		}
		v, err := nextToken(tokens, ignoreNewlines)
		if err != nil {
			return nil, err
		}
		if v.Type != "variable" {
			return nil, NewSyntaxErrorFromToken(
				v, "Unexpected token %#v. Expecting variable", v.Type)
		}
		vars = append(vars, &Var{Token: v})
	}
	return vars, nil
}

// VAR <variable> (, <variable>)*
func parseVar(start *Token, tokens *TokenSource) (Stmt, error) {
	vars, err := parseVarList(tokens, false)
	if err != nil {
		return nil, err
	}
	return &StmtVar{Token: start, Vars: vars}, nil
}

// LOOP { <statement>* }
func parseLoop(start *Token, tokens *TokenSource) (Stmt, error) {
	stmts, err := parseStatementBlock(tokens)
	if err != nil {
		return nil, err
	}
	return &StmtWhile{
		Token: start,
		Test:  &ExprBool{Token: start, Val: true},
		Body:  stmts,
	}, nil
}

// WHILE <expression> { <statement>* }
func parseWhile(start *Token, tokens *TokenSource) (Stmt, error) {
	expr, err := parseExpression(tokens, false)
	if err != nil {
		return nil, err
	}
	stmts, err := parseStatementBlock(tokens)
	if err != nil {
		return nil, err
	}
	return &StmtWhile{
		Token: start,
		Test:  expr,
		Body:  stmts,
	}, nil
}

// IMPORT <string> [WITHPREFIX <variable>]
func parseImport(start *Token, tokens *TokenSource) (Stmt, error) {
	module, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	if module.Type != "string" {
		return nil, NewSyntaxErrorFromToken(module,
			"Unexpected token %#v. Expecting module string", module.Type)
	}
	next, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	if next.Type != "keyword" || next.Val != "withprefix" {
		tokens.Push(next)
		return &StmtImport{
			Token: start,
			Path: &ExprString{
				Token: module,
				Val:   module.Val,
			}}, nil
	}
	prefix, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	if prefix.Type != "variable" {
		return nil, NewSyntaxErrorFromToken(prefix,
			"Unexpected token %#v. Expecting module prefix", prefix.Type)
	}
	return &StmtImport{
		Token: start,
		Path: &ExprString{
			Token: module,
			Val:   module.Val,
		},
		Prefix: &Var{
			Token: prefix,
		}}, nil
}

// UNIMPORT <string>
func parseUnimport(start *Token, tokens *TokenSource) (Stmt, error) {
	module, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	if module.Type != "string" {
		return nil, NewSyntaxErrorFromToken(module,
			"Unexpected token %#v. Expecting module string", module.Type)
	}
	return &StmtUnimport{
		Token: start,
		Path: &ExprString{
			Token: module,
			Val:   module.Val,
		}}, nil
}

// UNDEFINE <variable> (, <variable>)*
func parseUndefine(start *Token, tokens *TokenSource) (Stmt, error) {
	vars, err := parseVarList(tokens, false)
	if err != nil {
		return nil, err
	}
	return &StmtUndefine{Token: start, Vars: vars}, nil
}

// EXPORT <variable> (, <variable>)*
func parseExport(start *Token, tokens *TokenSource) (Stmt, error) {
	vars, err := parseVarList(tokens, false)
	if err != nil {
		return nil, err
	}
	return &StmtExport{Token: start, Vars: vars}, nil
}

// FUNC <variable> `(`[<variable> (, <variable>)*]`)` { <statement>* }
func parseFunc(start *Token, tokens *TokenSource) (Stmt, error) {
	name, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	if name.Type != "variable" {
		return nil, NewSyntaxErrorFromToken(
			name, "Unexpected token %#v. Expecting procedure name", name.Type)
	}
	leftparen, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	if leftparen.Type != "(" {
		return nil, NewSyntaxErrorFromToken(leftparen,
			"Unexpected token %#v. Expecting left parenthesis.", leftparen.Type)
	}

	rightparen, err := nextToken(tokens, true)
	if err != nil {
		return nil, err
	}
	var vars []*Var
	if rightparen.Type != ")" {
		if rightparen.Type != "variable" {
			return nil, NewSyntaxErrorFromToken(rightparen,
				"Unexpected token %#v. Expecting variable or \")\"", rightparen.Type)
		}
		tokens.Push(rightparen)
		vars, err = parseVarList(tokens, true)
		if err != nil {
			return nil, err
		}
		rightparen, err = nextToken(tokens, true)
		if err != nil {
			return nil, err
		}
		if rightparen.Type != ")" {
			return nil, NewSyntaxErrorFromToken(
				rightparen, "Unexpected token %#v. Expecting \")\"", rightparen.Type)
		}
	}
	stmts, err := parseStatementBlock(tokens)
	if err != nil {
		return nil, err
	}
	return &StmtFuncDef{
		Token: start,
		Name:  &Var{Token: name},
		Args:  vars,
		Body:  stmts,
	}, nil
}

// PROC <variable> [<variable> (, <variable>)*] { <statement>* }
func parseProc(start *Token, tokens *TokenSource) (Stmt, error) {
	name, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	if name.Type != "variable" {
		return nil, NewSyntaxErrorFromToken(
			name, "Unexpected token %#v. Expecting procedure name", name.Type)
	}
	leftbrace, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	var vars []*Var
	if leftbrace.Type != "{" {
		if leftbrace.Type != "variable" {
			return nil, NewSyntaxErrorFromToken(
				leftbrace, "Unexpected token %#v. Expecting variable or \"{\"",
				leftbrace.Type)
		}
		tokens.Push(leftbrace)
		vars, err = parseVarList(tokens, false)
		if err != nil {
			return nil, err
		}
		leftbrace, err = tokens.NextToken()
		if err != nil {
			return nil, err
		}
		if leftbrace.Type != "{" {
			return nil, NewSyntaxErrorFromToken(
				leftbrace, "Unexpected token %#v. Expecting \"{\"", leftbrace.Type)
		}
	}
	tokens.Push(leftbrace)
	stmts, err := parseStatementBlock(tokens)
	if err != nil {
		return nil, err
	}
	return &StmtProcDef{
		Token: start,
		Name:  &Var{Token: name},
		Args:  vars,
		Body:  stmts,
	}, nil
}

// RETURN <expression>
func parseReturn(start *Token, tokens *TokenSource) (Stmt, error) {
	expr, err := parseExpression(tokens, false)
	if err != nil {
		return nil, err
	}
	return &StmtReturn{
		Token: start,
		Val:   expr,
	}, nil
}

// <variable> = <expression>
func parseAssignment(lhs *Token, tokens *TokenSource) (Stmt, error) {
	expr, err := parseExpression(tokens, false)
	if err != nil {
		return nil, err
	}
	return &StmtAssignment{
		Token: lhs,
		Lhs:   &Var{Token: lhs},
		Rhs:   expr,
	}, nil
}

// <expression_order_2> [<expression> (, <expression>)* ]
func parseProcCall(tokens *TokenSource) (Stmt, error) {
	start, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	tokens.Push(start)
	proc, err := parseExprOrder2(tokens)
	if err != nil {
		return nil, err
	}
	tok, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	tokens.Push(tok)
	if tok.Type == "newline" || tok.Type == ";" || tok.Type == "}" {
		return &StmtProcCall{Token: start, Proc: proc}, nil
	}

	var args []Expr
	for {
		arg, err := parseExpression(tokens, false)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		tok, err := tokens.NextToken()
		if err != nil {
			return nil, err
		}
		if tok.Type != "," {
			tokens.Push(tok)
			break
		}
	}
	return &StmtProcCall{Token: start, Proc: proc, Args: args}, nil
}
