package main

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

type LineSource interface {
	NextLine() (lineno int, line string, err error)
}

type ReaderLineSource struct {
	buf    *bufio.Reader
	prompt func() error
	err    error
	lineno int
}

func NewReaderLineSource(r io.Reader, prompt func() error) *ReaderLineSource {
	return &ReaderLineSource{
		buf:    bufio.NewReader(r),
		prompt: prompt,
	}
}

func (ls *ReaderLineSource) NextLine() (lineno int, line string, err error) {
	if ls.err != nil {
		return 0, "", ls.err
	}
	if ls.prompt != nil {
		err = ls.prompt()
		if err != nil {
			ls.err = err
			return 0, "", err
		}
	}
	line, err = ls.buf.ReadString('\n')
	if err != nil {
		ls.err = err
		if err != io.EOF || len(line) == 0 {
			return 0, "", err
		}
	}
	ls.lineno += 1
	return ls.lineno, strings.TrimRightFunc(line, unicode.IsSpace), nil
}
