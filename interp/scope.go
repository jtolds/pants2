package interp

import (
	"fmt"

	"github.com/jtolds/pants2/ast"
)

type ModuleImporterFunc func(path string) (map[string]*ValueCell, error)

func (f ModuleImporterFunc) Import(path string) (map[string]*ValueCell, error) {
	return f(path)
}

type ModuleImporter interface {
	Import(path string) (map[string]*ValueCell, error)
}

type Scope interface {
	Flatten() Scope
	Fork() Scope

	Lookup(name string) *ValueCell
	Define(name string, v *ValueCell)
	Remove(name string)
	Export(stmt *ast.StmtExport) error
	Import(path, prefix string) error
	Unimport(path string) error
	Exports() map[string]*ValueCell
}

type ForkScope struct {
	parent Scope
	vars   map[string]*ValueCell
}

func NewForkScope(parent Scope) *ForkScope {
	var sf ForkScope
	sf.Init(parent)
	return &sf
}

func (f *ForkScope) Init(parent Scope) {
	f.parent = parent
	if f.vars != nil {
		for k := range f.vars {
			delete(f.vars, k)
		}
	}
}

func (f *ForkScope) Lookup(name string) *ValueCell {
	if f.vars != nil {
		if vc, exists := f.vars[name]; exists {
			return vc
		}
	}
	return f.parent.Lookup(name)
}

func (f *ForkScope) Define(name string, v *ValueCell) {
	if f.vars == nil {
		f.vars = map[string]*ValueCell{}
	}
	f.vars[name] = v
}

func (f *ForkScope) Export(stmt *ast.StmtExport) error {
	return NewRuntimeError(stmt.Token, "Unexpected export")
}

func (f *ForkScope) Exports() map[string]*ValueCell {
	return map[string]*ValueCell{}
}

func (f *ForkScope) Flatten() Scope {
	s := f.parent.Flatten()
	for k, v := range f.vars {
		if v != nil {
			s.Define(k, v)
		} else {
			s.Remove(k)
		}
	}
	return s
}

func (f *ForkScope) Fork() Scope {
	return NewForkScope(f)
}

func (f *ForkScope) Import(path, prefix string) error {
	return fmt.Errorf("TODO: import unsupported on forked scope")
}

func (f *ForkScope) Unimport(path string) error {
	return fmt.Errorf("TODO: unimport unsupported on forked scope")
}

func (f *ForkScope) Remove(name string) {
	f.Define(name, nil)
}

type FlatScope struct {
	vars      map[string]*ValueCell
	exports   map[string]*ValueCell
	importer  ModuleImporter
	unimports map[string]map[string]bool
}

func NewFlatScope(importer ModuleImporter) *FlatScope {
	return &FlatScope{
		vars:      map[string]*ValueCell{},
		importer:  importer,
		unimports: map[string]map[string]bool{},
	}
}

func (s *FlatScope) Lookup(name string) *ValueCell {
	return s.vars[name]
}

func (s *FlatScope) Define(name string, v *ValueCell) {
	s.vars[name] = v
}

func (s *FlatScope) Remove(name string) {
	delete(s.vars, name)
	for mod := range s.unimports {
		delete(s.unimports[mod], name)
	}
}

func (s *FlatScope) Export(stmt *ast.StmtExport) error {
	if s.exports == nil {
		return NewRuntimeError(stmt.Token, "Unexpected export")
	}
	for _, v := range stmt.Vars {
		if d, exists := s.exports[v.Token.Val]; exists {
			return NewRuntimeError(v.Token,
				"Exported variable \"%s\" already exported on file %#v, line %d",
				v.Token.Val, d.Def.Filename, d.Def.Lineno)
		}
	}
	for _, v := range stmt.Vars {
		cell := s.Lookup(v.Token.Val)
		if cell == nil {
			return NewRuntimeError(v.Token,
				"Variable %v not defined", v.Token.Val)
		}
		s.exports[v.Token.Val] = cell
	}
	return nil
}

func (s *FlatScope) Exports() map[string]*ValueCell {
	if s.exports == nil {
		s.exports = map[string]*ValueCell{}
	}
	return s.exports
}

func (s *FlatScope) copy() *FlatScope {
	c := &FlatScope{
		vars: make(map[string]*ValueCell, len(s.vars)),
		// deliberately don't copy exports
		importer:  s.importer,
		unimports: make(map[string]map[string]bool, len(s.unimports)),
	}
	for k, v := range s.vars {
		c.vars[k] = v
	}
	for mod, vars := range s.unimports {
		c.unimports[mod] = make(map[string]bool, len(vars))
		for k, v := range vars {
			c.unimports[mod][k] = v
		}
	}
	return c
}

func (s *FlatScope) Fork() Scope    { return NewForkScope(s) }
func (s *FlatScope) Flatten() Scope { return s.copy() }

func (s *FlatScope) Import(path, prefix string) error {
	if _, exists := s.unimports[path]; exists {
		return fmt.Errorf("%#v already imported", path)
	}
	vals, err := s.importer.Import(path)
	if err != nil {
		return err
	}
	if prefix != "" {
		prefix += "_"
	}
	for v := range vals {
		if d, exists := s.vars[prefix+v]; exists {
			return fmt.Errorf(
				"Export defines %#v, but %#v already defined on file %#v, line %d",
				prefix+v, prefix+v, d.Def.Filename, d.Def.Lineno)
		}
	}
	unimports := make(map[string]bool, len(vals))
	for v, cell := range vals {
		s.vars[prefix+v] = &ValueCell{
			Def: cell.Def,
			Val: cell.Val,
		}
		unimports[prefix+v] = true
	}
	s.unimports[path] = unimports
	return nil
}

func (s *FlatScope) Unimport(path string) error {
	vars, exists := s.unimports[path]
	if !exists {
		return fmt.Errorf("Module %#v not imported", path)
	}
	for v := range vars {
		delete(s.vars, v)
	}
	delete(s.unimports, path)
	return nil
}
