package tests

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/jtolds/pants2/app"
	"github.com/jtolds/pants2/interp"
	"github.com/jtolds/pants2/mods/std"
)

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertTrue(t *testing.T, val bool) {
	t.Helper()
	if !val {
		t.Fatal("expectation mismatch")
	}
}

func run(t *testing.T, code string, invals map[string]interp.Value) map[string]*interp.ValueCell {
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

func TestAdd(t *testing.T) {
	vals := run(t, `
		var x;
		x = 3;
		x = x + 1;
		export x;`, nil)

	assertTrue(t, len(vals) == 1)
	assertTrue(t, vals["x"].Val.(*interp.ValNumber).Val.Cmp(big.NewRat(4, 1)) == 0)
}

func TestAdd2(t *testing.T) {
	vals := run(t, `
		var x = 3;
		x = x + 1;
		export x;`, nil)

	assertTrue(t, len(vals) == 1)
	assertTrue(t, vals["x"].Val.(*interp.ValNumber).Val.Cmp(big.NewRat(4, 1)) == 0)
}

func TestSubproc(t *testing.T) {
	var testcalls []interp.Value
	testcall := func(args []interp.Value) error {
		if len(args) != 1 {
			return fmt.Errorf("expected 1 argument")
		}
		testcalls = append(testcalls, args[0])
		return nil
	}
	vals := run(t, `
		func y(a) { return a*2 }
		proc z a { testcall a*4 }

		var r1, r2, r3
		r1 = (100)*3
		r2 = y(r1)
		r3 = y((100)*5)
		z r1
		z (100)*5

		export r1, r2, r3`, map[string]interp.Value{"testcall": interp.ProcCB(testcall)})
	assertTrue(t, len(vals) == 3)
	assertTrue(t, vals["r1"].Val.(*interp.ValNumber).Val.Cmp(big.NewRat(300, 1)) == 0)
	assertTrue(t, vals["r2"].Val.(*interp.ValNumber).Val.Cmp(big.NewRat(600, 1)) == 0)
	assertTrue(t, vals["r3"].Val.(*interp.ValNumber).Val.Cmp(big.NewRat(1000, 1)) == 0)
	assertTrue(t, len(testcalls) == 2)
	assertTrue(t, testcalls[0].(*interp.ValNumber).Val.Cmp(big.NewRat(1200, 1)) == 0)
	assertTrue(t, testcalls[1].(*interp.ValNumber).Val.Cmp(big.NewRat(2000, 1)) == 0)
}
