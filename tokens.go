package main

import (
	"io"
	"unicode"
)

type Token struct {
	Lineno int
	Start  int
	Length int
	Type   string
	Repr   string
}

type Tokenizer struct {
	lineno  int
	chars   []rune
	charpos int
}

func NewTokenizer(lineno int, line string) *Tokenizer {
	return &Tokenizer{
		lineno: lineno,
		chars:  []rune(line),
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
			Lineno: t.lineno,
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
			Lineno: t.lineno,
			Start:  start,
			Length: t.charpos - start,
			Type:   string(t.chars[start:t.charpos])}, nil
	case '!':
		if t.charpos+1 >= len(t.chars) || t.chars[t.charpos+1] != '=' {
			return nil, NewSyntaxError(t.lineno, t.charpos, string(t.chars),
				"Unexpected exclamation point. Did you mean \"!=\"?")
		}
		t.charpos += 2
		return &Token{
			Lineno: t.lineno,
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
					return nil, NewSyntaxError(t.lineno, t.charpos,
						string(t.chars), "Unexpected second decimal point")
				}
				decimal = true
			}
			t.charpos += 1
		}
		if decimal && t.charpos-1 == start {
			return nil, NewSyntaxError(t.lineno, t.charpos-1,
				string(t.chars), "Number expected before or after decimal point")
		}
		if t.charpos < len(t.chars) && unicode.IsLetter(t.chars[t.charpos]) {
			return nil, NewSyntaxError(t.lineno, t.charpos,
				string(t.chars), "Unexpected letter after number")
		}
		if t.charpos < len(t.chars) && t.chars[t.charpos] == '_' {
			return nil, NewSyntaxError(t.lineno, t.charpos,
				string(t.chars), "Unexpected underscore after number")
		}
		return &Token{
			Lineno: t.lineno,
			Start:  start,
			Length: t.charpos - start,
			Type:   "number",
			Repr:   string(t.chars[start:t.charpos])}, nil
	}

	if unicode.IsLetter(t.chars[t.charpos]) || t.chars[t.charpos] == '_' {
		start := t.charpos
		for t.charpos < len(t.chars) &&
			(unicode.IsLetter(t.chars[t.charpos]) ||
				unicode.IsNumber(t.chars[t.charpos]) ||
				map[rune]bool{'.': true, '_': true}[t.chars[t.charpos]]) {
			t.charpos += 1
		}
		return &Token{
			Lineno: t.lineno,
			Start:  start,
			Length: t.charpos - start,
			Type:   "variable",
			Repr:   string(t.chars[start:t.charpos])}, nil
	}

	return nil, NewSyntaxError(t.lineno, t.charpos, string(t.chars),
		"Unexpected character: %#v", string(t.chars[t.charpos]))
}

func (t *Tokenizer) parseString() (*Token, error) {
	if t.chars[t.charpos] != '"' {
		return nil, NewSyntaxError(t.lineno, t.charpos, string(t.chars),
			"String expected. Found %#v instead.", string(t.chars[t.charpos]))
	}
	start := t.charpos
	t.charpos += 1
	for ; t.charpos < len(t.chars); t.charpos += 1 {
		if t.chars[t.charpos] == '\\' {
			t.charpos += 1
			if t.charpos >= len(t.chars) {
				break
			}
			switch t.chars[t.charpos] {
			case '\\', '"', 'n', 't', 'x':
			default:
				return nil, NewSyntaxError(t.lineno, t.charpos-1, string(t.chars),
					"String escape value unknown: \\%v.\n"+
						"Expected one of \\\\, \\\", \\n, \\t, or \\x",
					string(t.chars[t.charpos]))
			}
			continue
		}
		if t.chars[t.charpos] == '"' {
			t.charpos += 1
			return &Token{
				Lineno: t.lineno,
				Start:  start,
				Length: t.charpos - start,
				Type:   "string",
				Repr:   string(t.chars[start+1 : t.charpos-1]),
			}, nil
		}
	}
	return nil, NewSyntaxError(t.lineno, start, string(t.chars),
		"String started but not ended.")
}

func (t *Tokenizer) skipWhitespace() {
	for t.charpos < len(t.chars) && unicode.IsSpace(t.chars[t.charpos]) {
		t.charpos += 1
	}
}

func Tokenize(lineno int, line string) (rv []*Token, err error) {
	tok := NewTokenizer(lineno, line)
	for {
		t, err := tok.Next()
		if t != nil {
			rv = append(rv, t)
		}
		if err != nil {
			if err == io.EOF {
				return rv, nil
			}
			return rv, err
		}
	}
}
