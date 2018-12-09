package interp

import (
	"fmt"
)

type ModuleImporterFunc func(path string) (map[string]*ValueCell, error)

func (f ModuleImporterFunc) Import(path string) (map[string]*ValueCell, error) {
	return f(path)
}

type ModuleImporter interface {
	Import(path string) (map[string]*ValueCell, error)
}

type Scope struct {
	vars      map[string]*ValueCell
	exports   map[string]*ValueCell
	importer  ModuleImporter
	unimports map[string]map[string]bool
}

func NewScope(importer ModuleImporter) *Scope {
	return &Scope{
		vars:      map[string]*ValueCell{},
		importer:  importer,
		unimports: map[string]map[string]bool{},
	}
}

func (s *Scope) Exports() map[string]*ValueCell {
	if s.exports == nil {
		s.exports = map[string]*ValueCell{}
	}
	return s.exports
}

func (s *Scope) Copy() *Scope {
	c := &Scope{
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

func (s *Scope) Import(path string, vals map[string]*ValueCell, prefix string) (
	err error) {
	if _, exists := s.unimports[path]; exists {
		return fmt.Errorf("%#v already imported", path)
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

func (s *Scope) Unimport(path string) error {
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
