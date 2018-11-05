package ast

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

type Line struct {
	Filename string
	Lineno   int
	Line     string
}

type LineSource interface {
	Pos() (filename string, lineno int)
	NextLine() (line *Line, err error)
}

type ReaderLineSource struct {
	filename string
	buf      *bufio.Reader
	prompt   func() error
	err      error
	lineno   int
}

func NewReaderLineSource(filename string, r io.Reader, prompt func() error) (
	ls *ReaderLineSource) {
	return &ReaderLineSource{
		filename: filename,
		buf:      bufio.NewReader(r),
		prompt:   prompt,
	}
}

func (ls *ReaderLineSource) NextLine() (*Line, error) {
	if ls.err != nil {
		return nil, ls.err
	}
	if ls.prompt != nil {
		err := ls.prompt()
		if err != nil {
			ls.err = err
			return nil, err
		}
	}
	line, err := ls.buf.ReadString('\n')
	if err != nil {
		ls.err = err
		if err != io.EOF || len(line) == 0 {
			return nil, err
		}
	}
	ls.lineno += 1
	return &Line{
		Filename: ls.filename,
		Lineno:   ls.lineno,
		Line:     strings.TrimRightFunc(line, unicode.IsSpace)}, nil
}

func (ls *ReaderLineSource) Pos() (string, int) {
	return ls.filename, ls.lineno
}
