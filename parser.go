package main

type Parser struct {
	filename string
}

func NewParser(filename string) *Parser {
	return &Parser{filename: filename}
}

func (p *Parser) ParseNext(lines LineSource) (Stmt, error) {
	lineno, line, err := lines.NextLine()
	if err != nil {
		return nil, err
	}

	tokens, err := Tokenize(lineno, line)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, nil
	}

	proc := tokens[0]
	if proc.Type != "variable" {
		return nil, NewSyntaxErrorFromToken(proc, line,
			"Unexpected token %#v. Expecting procedure.", proc.Type)
	}

	switch proc.Repr {
	case "if", "IF": // IF <expression> { <statement>* }
	case "else", "ELSE": // ELSE { <statement>* }
	case "var", "VAR": // VAR <variable> (, <variable>)*
	case "loop", "LOOP": // LOOP { <statement>* }
	case "while", "WHILE": // WHILE <expression> { <statement>* }
	case "import", "IMPORT": // IMPORT <string> [WITH PREFIX <variable>]
	case "unimport", "UNIMPORT": // UNIMPORT <string>
	case "undefine", "UNDEFINE": // UNDEFINE <variable> (, <variable>)*
	case "export", "EXPORT": // EXPORT <variable> (, <variable>)*
	case "func", "FUNC": // FUNC <variable> `(`[<variable> (, <variable>)*]`)` { <statement>* }
	case "proc", "PROC": // PROC <variable> [<variable> (, <variable>)*] { <statement>* }
	case "break", "BREAK", "next", "NEXT", "done", "DONE":
	case "return", "RETURN": // RETURN <expression>
	default:
		// <variable> = <expression>
		// <expression> [<expression> (, <expression>)*]
	}

	panic("unreachable")
}
