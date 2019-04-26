package main

import (
	"flag"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/jtolds/pants2/app"
	"github.com/jtolds/pants2/mods/std"
	"github.com/jtolds/pants2/mods/vis2d"
)

var (
	cpuProfile = flag.String("cpu-profile", "", "profile output file")
	memProfile = flag.String("mem-profile", "", "profile output file")
)

func main() {
	err := Main()
	if err != nil {
		panic(err)
	}
}

func Main() error {
	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			return err
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	a := app.NewApp()
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

	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			return err
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
			return err
		}
		f.Close()
	}

	return apperr
}
