package main

import (
	"fmt"
)

type SyntaxError struct {
	lineno  int
	charpos int
	line    string
	msg     string
}

func NewSyntaxError(lineno, charpos int, line string,
	format string, args ...interface{}) *SyntaxError {
	return &SyntaxError{
		lineno:  lineno,
		charpos: charpos,
		line:    line,
		msg:     fmt.Sprintf(format, args...),
	}
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("Syntax error on line %d, character %d: %s",
		e.lineno, e.charpos+1, e.msg)
}

func IsSyntaxError(err error) bool {
	_, ok := err.(*SyntaxError)
	return ok
}

func IsHandledError(err error) bool {
	return IsSyntaxError(err)
}
