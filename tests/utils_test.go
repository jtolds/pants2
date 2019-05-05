package tests

import (
	"bytes"
	"testing"

	"github.com/jtolds/pants2/app"
	"github.com/jtolds/pants2/interp"
	"github.com/jtolds/pants2/lib/big"
	"github.com/jtolds/pants2/mods/std"
)

func assertNoErr(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertTrue(t testing.TB, val bool) {
	t.Helper()
	if !val {
		t.Fatal("expectation mismatch")
	}
}

func run(t testing.TB, code string, invals map[string]interp.Value) map[string]*interp.ValueCell {
	a := app.NewApp()
	a.DefineModule("std", std.Mod)
	a.RunInDefaultScope(`import "std";`)
	if invals != nil {
		a.DefineModule("_test", func() (map[string]interp.Value, error) { return invals, nil })
		a.RunInDefaultScope(`import "_test";`)
	}
	vals, err := a.Load("test", bytes.NewReader([]byte(code)))
	assertNoErr(t, err)
	return vals
}

func assertNumEqual(t testing.TB, arg interp.Value, expected *big.Rat) {
	t.Helper()
	v := arg.(interp.ValNumber).Val
	assertTrue(t, v.Cmp(expected) == 0)
}
