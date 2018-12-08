package main

import (
	"flag"
	"os"

	"github.com/jtolds/pants2/mods/std"
	"github.com/jtolds/pants2/mods/vis2d"
)

func main() {
	err := Main()
	if err != nil {
		panic(err)
	}
}

func Main() error {
	flag.Parse()
	a := NewApp()
	a.DefineModule("std", std.Mod)
	a.DefineModule("vis2d", vis2d.Mod)
	err := a.RunInDefaultScope(`import "vis2d"; import "std";`)
	if err != nil {
		return err
	}
	var apperr error
	go func() {
		defer vis2d.Stop()
		file := flag.Arg(0)
		if file != "" {
			_, apperr = a.LoadFile(file)
			return
		}
		_, apperr = a.LoadInteractive(os.Stdin, os.Stderr)
	}()
	vis2d.Run()
	return apperr
}
