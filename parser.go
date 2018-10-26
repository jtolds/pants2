package main

import (
	"io"
	"strconv"
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
	case "int":
		val, err := strconv.ParseInt(tok.Val, 10, 64)
		if err != nil {
			return nil, err
		}
		return &ExprInt{Token: tok, Val: val}, nil
	case "float":
		val, err := strconv.ParseFloat(tok.Val, 64)
		if err != nil {
			return nil, err
		}
		return &ExprFloat{Token: tok, Val: val}, nil
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

func parseExpression(tokens *TokenSource, ignoreNewlines bool) (Expr, error) {
	val, err := parseExprOrder3(tokens, ignoreNewlines)
	if err != nil {
		return nil, err
	}
	op, err := nextToken(tokens, ignoreNewlines)
	if err != nil {
		return nil, err
	}
	switch op.Type {
	// TODO: order of operations
	// TODO: left-associative
	case "==", "!=", "<", "<=", ">", ">=", "+", "-", "*", "%", "/", "and", "or":
		end, err := parseExpression(tokens, ignoreNewlines)
		if err != nil {
			return nil, err
		}
		return &ExprOp{
			Token: op,
			Left:  val,
			Op:    op,
			Right: end,
		}, nil
	default:
		tokens.Push(op)
		return val, err
	}
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

// VAR <variable> (, <variable>)*
func parseVar(start *Token, tokens *TokenSource) (Stmt, error) {
	panic("TODO")
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

// IMPORT <string> [WITH PREFIX <variable>]
func parseImport(start *Token, tokens *TokenSource) (Stmt, error) {
	panic("TODO")
}

// UNIMPORT <string>
func parseUnimport(start *Token, tokens *TokenSource) (Stmt, error) {
	panic("TODO")
}

// UNDEFINE <variable> (, <variable>)*
func parseUndefine(start *Token, tokens *TokenSource) (Stmt, error) {
	panic("TODO")
}

// EXPORT <variable> (, <variable>)*
func parseExport(start *Token, tokens *TokenSource) (Stmt, error) {
	panic("TODO")
}

// FUNC <variable> `(`[<variable> (, <variable>)*]`)` { <statement>* }
func parseFunc(start *Token, tokens *TokenSource) (Stmt, error) {
	panic("TODO")
}

// PROC <variable> [<variable> (, <variable>)*] { <statement>* }
func parseProc(start *Token, tokens *TokenSource) (Stmt, error) {
	panic("TODO")
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
	proc, err := parseExprOrder2(tokens)
	if err != nil {
		return nil, err
	}
	tok, err := tokens.NextToken()
	if err != nil {
		return nil, err
	}
	if tok.Type == "newline" || tok.Type == ";" {
		return &StmtProcCall{Proc: proc}, nil
	}
	tokens.Push(tok)

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
	return &StmtProcCall{Proc: proc, Args: args}, nil
}

// `(`<expression>`)` [<expression> (, <expression>)* ]
func parseExprProcCall(proc *Token, tokens *TokenSource) (Stmt, error) {
	panic("TODO")
}
