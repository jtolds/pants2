package main

import (
	"fmt"
	"io"

	"github.com/jtolds/pants2/ast"
	"github.com/jtolds/pants2/interp"
)

type App struct {
	defaultScope *interp.Scope
}

func NewApp() (a *App) {
	a = &App{}
	a.defaultScope = interp.NewScope(interp.ModuleImporterFunc(a.importMod))
	return a
}

func (a *App) Load(name string, input io.Reader) error {
	s := a.defaultScope.Copy()
	s.EnableExports()
	tokens := ast.NewTokenSource(ast.NewReaderLineSource(name, input, nil))
	for {
		stmt, err := ast.ParseStatement(tokens)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		err = s.Run(stmt)
		if err != nil {
			return err
		}
	}
}

func (a *App) LoadInteractive(input io.Reader, output io.Writer) error {
	s := a.defaultScope.Copy()
	s.EnableExports()
	tokens := ast.NewTokenSource(ast.NewReaderLineSource("<stdin>", input,
		func() error {
			_, err := fmt.Fprintf(output, "> ")
			return err
		}))
	for {
		stmt, err := ast.ParseStatement(tokens)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			if !interp.IsHandledError(err) {
				return err
			}
			_, err = fmt.Fprintln(output, err)
			if err != nil {
				return err
			}
			tokens.ResetLine()
			continue
		}
		err = s.Run(stmt)
		if err != nil {
			if !interp.IsHandledError(err) {
				return err
			}
			_, err = fmt.Fprintln(output, err)
			if err != nil {
				return err
			}
			tokens.ResetLine()
		}
	}
}

func (a *App) Define(name string, value interp.Value) {
	a.defaultScope.Define(name, value)
}

func (a *App) importMod(path string) (map[string]*interp.ValueCell, error) {
	panic("TODO")
}
