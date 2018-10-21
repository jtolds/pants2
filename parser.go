package main

import (
	"fmt"
)

type Parser struct {
	filename string
}

func NewParser(filename string) *Parser {
	return &Parser{filename: filename}
}

func (p *Parser) ParseNext(lines LineSource) error {
	lineno, line, err := lines.NextLine()
	if err != nil {
		return err
	}

	tokens, err := Tokenize(lineno, line)
	if err != nil {
		return err
	}

	for _, token := range tokens {
		_, err = fmt.Printf("%#v\n", token)
	}
	return err
}
