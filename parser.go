package main

func ParseStatement(tokens *TokenSource) (stmt Stmt, err error) {
	var token *Token
	for {
		token, err = tokens.NextToken()
		if err != nil {
			return nil, err
		}

		if token.Type != "newline" && token.Type != ";" {
			break
		}
	}

	if token.Type == "keyword" {
		switch token.Repr {
		case "if":
			return parseIf(token, tokens)
		case "else":
			return parseElse(token, tokens)
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
		return parseVarProcCall(token, tokens)
	}

	if token.Type == "(" {
		return parseExprProcCall(token, tokens)
	}

	return nil, NewSyntaxErrorFromToken(token,
		"Unexpected token %#v. Expecting statement.", token.Type)
}

func parseExpression(tokens *TokenSource) (Expr, error) {
	panic("TODO")
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
		rightbrace, err := tokens.NextToken()
		if err != nil {
			return nil, err
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

// IF <expression> { <statement>* }
func parseIf(start *Token, tokens *TokenSource) (Stmt, error) {
	expr, err := parseExpression(tokens)
	if err != nil {
		return nil, err
	}
	stmts, err := parseStatementBlock(tokens)
	if err != nil {
		return nil, err
	}
	return &StmtIf{
		Token: start,
		Test:  expr,
		Body:  stmts,
	}, nil
}

// ELSE { <statement>* }
func parseElse(start *Token, tokens *TokenSource) (Stmt, error) {
	stmts, err := parseStatementBlock(tokens)
	if err != nil {
		return nil, err
	}
	return &StmtElse{
		Token: start,
		Body:  stmts,
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
	expr, err := parseExpression(tokens)
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
	expr, err := parseExpression(tokens)
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
	expr, err := parseExpression(tokens)
	if err != nil {
		return nil, err
	}
	return &StmtAssignment{
		Token: lhs,
		Lhs:   &Var{Token: lhs},
		Rhs:   expr,
	}, nil
}

// <variable> [<expression> (, <expression>)* ]
func parseVarProcCall(proc *Token, tokens *TokenSource) (Stmt, error) {
	panic("TODO")
}

// `(`<expression>`)` [<expression> (, <expression>)* ]
func parseExprProcCall(proc *Token, tokens *TokenSource) (Stmt, error) {
	panic("TODO")
}
