package main

import (
	"io"
	"unicode"
)

type Token struct {
	lineno int
	start  int
	length int
	typ    string
	repr   string
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
	case ',', '{', '}', '[', ']', '(', ')', ';':
		t.charpos += 1
		return &Token{
			lineno: t.lineno,
			start:  t.charpos - 1,
			length: 1,
			typ:    string(t.chars[t.charpos-1])}, nil
	case '"':
		return t.parseString()
	case '=', '<', '>':
		start := t.charpos
		t.charpos += 1
		if t.charpos < len(t.chars) && t.chars[t.charpos] == '=' {
			t.charpos += 1
		}
		return &Token{
			lineno: t.lineno,
			start:  start,
			length: t.charpos - start,
			typ:    string(t.chars[start:t.charpos])}, nil
	case '!':
		if t.charpos+1 >= len(t.chars) || t.chars[t.charpos+1] != '=' {
			return nil, NewSyntaxError(t.lineno, t.charpos, string(t.chars),
				"Unexpected exclamation point. Did you mean \"!=\"?")
		}
		t.charpos += 2
		return &Token{
			lineno: t.lineno,
			start:  t.charpos - 2,
			length: 2,
			typ:    "!="}, nil
	}

	start := t.charpos
	for t.charpos < len(t.chars) &&
		(unicode.IsLetter(t.chars[t.charpos]) ||
			unicode.IsNumber(t.chars[t.charpos])) {
		t.charpos += 1
	}
	if t.charpos > start {
		return &Token{
			lineno: t.lineno,
			start:  start,
			length: t.charpos - start,
			typ:    "alphanum",
			repr:   string(t.chars[start:t.charpos])}, nil
	}

	return nil, NewSyntaxError(t.lineno, start, string(t.chars),
		"Unexpected character: %#v", string(t.chars[start]))
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
					"String escape value unknown: %#v.\nExpected \\\\ or \\\"",
					string(t.chars[t.charpos]))
			}
			continue
		}
		if t.chars[t.charpos] == '"' {
			t.charpos += 1
			return &Token{
				lineno: t.lineno,
				start:  start,
				length: t.charpos - start,
				typ:    "string",
				repr:   string(t.chars[start+1 : t.charpos-1]),
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
