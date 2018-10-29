package main

import (
	"fmt"
)

type SyntaxError struct {
	line    *Line
	charpos int
	msg     string
}

func NewSyntaxError(line *Line, charpos int,
	format string, args ...interface{}) *SyntaxError {
	return &SyntaxError{
		line:    line,
		charpos: charpos,
		msg:     fmt.Sprintf(format, args...),
	}
}

func NewSyntaxErrorFromToken(token *Token,
	format string, args ...interface{}) *SyntaxError {
	return NewSyntaxError(token.Line, token.Start, format, args...)
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("Syntax error on file %#v, line %d, character %d: %s",
		e.line.Filename, e.line.Lineno, e.charpos+1, e.msg)
}

func IsSyntaxError(err error) bool {
	_, ok := err.(*SyntaxError)
	return ok
}

func IsHandledError(err error) bool {
	return IsSyntaxError(err) || IsRuntimeError(err)
}

type RuntimeError struct {
	token *Token
	msg   string
}

func NewRuntimeError(token *Token, format string, args ...interface{}) (
	re *RuntimeError) {
	return &RuntimeError{
		token: token,
		msg:   fmt.Sprintf(format, args...),
	}
}

func IsRuntimeError(err error) bool {
	_, ok := err.(*RuntimeError)
	return ok
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error on file %#v, line %d: %s",
		e.token.Line.Filename, e.token.Line.Lineno, e.msg)
}
