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
	return IsSyntaxError(err) || IsRuntimeError(err) || IsControlError(err)
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

type ControlType string

var (
	CtrlBreak  ControlType = "break"
	CtrlNext   ControlType = "next"
	CtrlDone   ControlType = "done"
	CtrlReturn ControlType = "return"
)

type ControlError struct {
	token *Token
	typ   ControlType
	val   Value
}

func NewControlError(token *Token, value Value) *ControlError {
	if token.Type != "keyword" {
		panic(fmt.Sprintf("unexpected token type %s", token.Type))
	}
	switch ControlType(token.Val) {
	case CtrlBreak, CtrlNext, CtrlDone:
		if value != nil {
			panic("unexpected value")
		}
	case CtrlReturn:
		if value == nil {
			panic("expected value")
		}
	default:
		panic(fmt.Sprintf("unexpected keyword %s", token.Val))
	}
	return &ControlError{
		token: token,
		typ:   ControlType(token.Val),
		val:   value,
	}
}

func (e *ControlError) Error() string {
	return fmt.Sprintf("Unexpected \"%s\" on file %#v, line %d",
		string(e.typ), e.token.Line.Filename, e.token.Line.Lineno)
}

func IsControlError(err error) bool {
	_, ok := err.(*ControlError)
	return ok
}

func IsControlErrorType(err error, typ ControlType) bool {
	e, ok := err.(*ControlError)
	if !ok {
		return false
	}
	return e.typ == typ
}
