package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jtolds/pants2/ast"
	"github.com/jtolds/pants2/interp"
)

type App struct {
	defaultScope *interp.Scope
	modules      map[string]map[string]*interp.ValueCell
}

func NewApp() (a *App) {
	a = &App{
		modules: map[string]map[string]*interp.ValueCell{},
	}
	a.defaultScope = interp.NewScope(interp.ModuleImporterFunc(a.importMod))
	return a
}

func (a *App) Load(name string, input io.Reader) (
	map[string]*interp.ValueCell, error) {
	s := a.defaultScope.Copy()
	rv := s.Exports()
	tokens := ast.NewTokenSource(ast.NewReaderLineSource(name, input, nil))
	for {
		stmt, err := ast.ParseStatement(tokens)
		if err != nil {
			if err == io.EOF {
				a.modules[name] = rv
				return rv, nil
			}
			return nil, err
		}
		err = s.Run(stmt)
		if err != nil {
			return nil, err
		}
	}
}

func (a *App) LoadInteractive(input io.Reader, output io.Writer) (
	map[string]*interp.ValueCell, error) {
	s := a.defaultScope.Copy()
	rv := s.Exports()
	tokens := ast.NewTokenSource(ast.NewReaderLineSource("<stdin>", input,
		func() error {
			_, err := fmt.Fprintf(output, "> ")
			return err
		}))
	for {
		stmt, err := ast.ParseStatement(tokens)
		if err != nil {
			if err == io.EOF {
				return rv, nil
			}
			if !interp.IsHandledError(err) {
				return nil, err
			}
			_, err = fmt.Fprintln(output, err)
			if err != nil {
				return nil, err
			}
			tokens.ResetLine()
			continue
		}
		err = s.Run(stmt)
		if err != nil {
			if !interp.IsHandledError(err) {
				return nil, err
			}
			_, err = fmt.Fprintln(output, err)
			if err != nil {
				return nil, err
			}
			tokens.ResetLine()
		}
	}
}

func (a *App) Define(name string, value interp.Value) {
	a.defaultScope.Define(name, value)
}

func (a *App) importMod(path string) (map[string]*interp.ValueCell, error) {
	rv, exists := a.modules[path]
	if exists {
		if rv == nil {
			// TODO
			return nil, fmt.Errorf("import cycle detected")
		}
		return rv, nil
	}
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	a.modules[path] = nil
	return a.Load(path, fh)
}
