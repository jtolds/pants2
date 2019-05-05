package tests

import (
	"fmt"
	"testing"

	"github.com/jtolds/pants2/interp"
	"github.com/jtolds/pants2/lib/big"
)

func TestAdd(t *testing.T) {
	vals := run(t, `
		var x;
		x = 3;
		x = x + 1;
		export x;`, nil)

	assertTrue(t, len(vals) == 1)
	assertNumEqual(t, vals["x"].Val, big.NewRat(4, 1))
}

func TestAdd2(t *testing.T) {
	vals := run(t, `
		var x = 3;
		x = x + 1;
		export x;`, nil)

	assertTrue(t, len(vals) == 1)
	assertNumEqual(t, vals["x"].Val, big.NewRat(4, 1))
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
	assertNumEqual(t, vals["r1"].Val, big.NewRat(300, 1))
	assertNumEqual(t, vals["r2"].Val, big.NewRat(600, 1))
	assertNumEqual(t, vals["r3"].Val, big.NewRat(1000, 1))
	assertTrue(t, len(testcalls) == 2)
	assertNumEqual(t, testcalls[0], big.NewRat(1200, 1))
	assertNumEqual(t, testcalls[1], big.NewRat(2000, 1))
}
