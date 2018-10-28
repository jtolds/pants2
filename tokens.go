package main

import (
	"io"
	"strings"
	"unicode"
)

type Token struct {
	Line   *Line
	Start  int
	Length int
	Type   string
	Val    string
}

type Tokenizer struct {
	line    *Line
	chars   []rune
	charpos int
}

func NewTokenizer(line *Line) *Tokenizer {
	return &Tokenizer{
		line:  line,
		chars: []rune(line.Line),
	}
}

func (t *Tokenizer) Next() (*Token, error) {
	t.skipWhitespace()
	if t.charpos >= len(t.chars) || t.chars[t.charpos] == '#' {
		return nil, io.EOF
	}

	switch t.chars[t.charpos] {
	case ',', '{', '}', '[', ']', '(', ')', ';', '+', '-', '*', '/', '%':
		t.charpos += 1
		return &Token{
			Line:   t.line,
			Start:  t.charpos - 1,
			Length: 1,
			Type:   string(t.chars[t.charpos-1])}, nil
	case '"':
		return t.parseString()
	case '=', '<', '>':
		start := t.charpos
		t.charpos += 1
		if t.charpos < len(t.chars) && t.chars[t.charpos] == '=' {
			t.charpos += 1
		}
		return &Token{
			Line:   t.line,
			Start:  start,
			Length: t.charpos - start,
			Type:   string(t.chars[start:t.charpos])}, nil
	case '!':
		if t.charpos+1 >= len(t.chars) || t.chars[t.charpos+1] != '=' {
			return nil, NewSyntaxError(t.line, t.charpos,
				"Unexpected exclamation point. Did you mean \"!=\"?")
		}
		t.charpos += 2
		return &Token{
			Line:   t.line,
			Start:  t.charpos - 2,
			Length: 2,
			Type:   "!="}, nil
	}

	if unicode.IsNumber(t.chars[t.charpos]) || t.chars[t.charpos] == '.' {
		start := t.charpos
		decimal := false
		for t.charpos < len(t.chars) && (unicode.IsNumber(t.chars[t.charpos]) ||
			t.chars[t.charpos] == '.') {
			if t.chars[t.charpos] == '.' {
				if decimal {
					return nil, NewSyntaxError(t.line, t.charpos,
						"Unexpected second decimal point")
				}
				decimal = true
			}
			t.charpos += 1
		}
		if decimal && t.charpos-1 == start {
			return nil, NewSyntaxError(t.line, t.charpos-1,
				"Number expected before or after decimal point")
		}
		if t.charpos < len(t.chars) && unicode.IsLetter(t.chars[t.charpos]) {
			return nil, NewSyntaxError(t.line, t.charpos,
				"Unexpected letter after number")
		}
		if t.charpos < len(t.chars) && t.chars[t.charpos] == '_' {
			return nil, NewSyntaxError(t.line, t.charpos,
				"Unexpected underscore after number")
		}
		typ := "int"
		if decimal {
			typ = "float"
		}
		return &Token{
			Line:   t.line,
			Start:  start,
			Length: t.charpos - start,
			Type:   typ,
			Val:    string(t.chars[start:t.charpos])}, nil
	}

	if unicode.IsLetter(t.chars[t.charpos]) || t.chars[t.charpos] == '_' {
		start := t.charpos
		for t.charpos < len(t.chars) &&
			(unicode.IsLetter(t.chars[t.charpos]) ||
				unicode.IsNumber(t.chars[t.charpos]) ||
				map[rune]bool{'.': true, '_': true}[t.chars[t.charpos]]) {
			t.charpos += 1
		}
		name := string(t.chars[start:t.charpos])
		switch name {
		case "if", "IF", "else", "ELSE", "var", "VAR", "loop", "LOOP",
			"while", "WHILE", "import", "IMPORT", "unimport", "UNIMPORT",
			"undefine", "UNDEFINE", "export", "EXPORT", "func", "FUNC",
			"proc", "PROC", "break", "BREAK", "next", "NEXT", "done", "DONE",
			"return", "RETURN", "withprefix", "WITHPREFIX":
			return &Token{
				Line:   t.line,
				Start:  start,
				Length: t.charpos - start,
				Type:   "keyword",
				Val:    strings.ToLower(name),
			}, nil
		case "true", "TRUE", "false", "FALSE":
			return &Token{
				Line:   t.line,
				Start:  start,
				Length: t.charpos - start,
				Type:   "bool",
				Val:    strings.ToLower(name),
			}, nil
		case "and", "AND", "or", "OR", "not", "NOT":
			return &Token{
				Line:   t.line,
				Start:  start,
				Length: t.charpos - start,
				Type:   strings.ToLower(name),
			}, nil
		default:
			return &Token{
				Line:   t.line,
				Start:  start,
				Length: t.charpos - start,
				Type:   "variable",
				Val:    name,
			}, nil
		}
	}

	return nil, NewSyntaxError(t.line, t.charpos,
		"Unexpected character: %#v", string(t.chars[t.charpos]))
}

func (t *Tokenizer) parseString() (*Token, error) {
	if t.chars[t.charpos] != '"' {
		return nil, NewSyntaxError(t.line, t.charpos,
			"String expected. Found %#v instead.", string(t.chars[t.charpos]))
	}
	start := t.charpos
	t.charpos += 1
	value := make([]rune, 0, len(t.chars)-t.charpos)
	for ; t.charpos < len(t.chars); t.charpos += 1 {
		if t.chars[t.charpos] == '\\' {
			t.charpos += 1
			if t.charpos >= len(t.chars) {
				break
			}
			switch t.chars[t.charpos] {
			case '\\':
				value = append(value, '\\')
			case '"':
				value = append(value, '"')
			case 'n':
				value = append(value, '\n')
			case 't':
				value = append(value, '\t')
			default:
				return nil, NewSyntaxError(t.line, t.charpos-1,
					"String escape value unknown: \\%v.\n"+
						"Expected one of \\\\, \\\", \\n, \\t",
					string(t.chars[t.charpos]))
			}
			continue
		} else if t.chars[t.charpos] == '"' {
			t.charpos += 1
			return &Token{
				Line:   t.line,
				Start:  start,
				Length: t.charpos - start,
				Type:   "string",
				Val:    string(value),
			}, nil
		} else {
			value = append(value, t.chars[t.charpos])
		}
	}
	return nil, NewSyntaxError(t.line, start,
		"String started but not ended.")
}

func (t *Tokenizer) skipWhitespace() {
	for t.charpos < len(t.chars) && unicode.IsSpace(t.chars[t.charpos]) {
		t.charpos += 1
	}
}

func Tokenize(line *Line) (rv []*Token, err error) {
	tok := NewTokenizer(line)
	for {
		t, err := tok.Next()
		if t != nil {
			rv = append(rv, t)
		}
		if err != nil {
			if err == io.EOF {
				rv = append(rv, &Token{
					Line:   line,
					Start:  len(line.Line),
					Length: 1,
					Type:   "newline",
				})
				return rv, nil
			}
			return rv, err
		}
	}
}

type TokenSource struct {
	ls     LineSource
	tokens []*Token
	pushed []*Token
	end    bool
}

func NewTokenSource(ls LineSource) *TokenSource {
	return &TokenSource{ls: ls}
}

func (t *TokenSource) NextToken() (rv *Token, err error) {
	if t.end {
		return nil, io.EOF
	}
	if len(t.pushed) > 0 {
		last := len(t.pushed) - 1
		rv, t.pushed = t.pushed[last], t.pushed[:last]
		return rv, nil
	}
	if len(t.tokens) == 0 {
		line, err := t.ls.NextLine()
		if err != nil {
			if err == io.EOF {
				filename, lineno := t.ls.Pos()
				t.end = true
				return &Token{
					Line: &Line{
						Filename: filename,
						Lineno:   lineno,
					},
					Type: "eof",
				}, nil
			}
			return nil, err
		}
		tokens, err := Tokenize(line)
		if err != nil {
			return nil, err
		}
		t.tokens = tokens
	}
	rv, t.tokens = t.tokens[0], t.tokens[1:]
	return rv, nil
}

func (t *TokenSource) Push(tok *Token) {
	t.pushed = append(t.pushed, tok)
}
