package tests

import (
	"testing"
)

func BenchmarkLines(b *testing.B) {
	for n := 0; n < b.N; n++ {
		run(b, `
func intdiv(x, y) {
  var m = x % y
  m = x - m
  return m / y
}

proc noop x, y {}

proc line x1, y1, x2, y2 {
  if x2 < x1 {
    var x2t = x2; x2 = x1; x1 = x2t
    var y2t = y2; y2 = y1; y1 = y2t
  }

  var y = y1, ydir = 1
  if y2 < y1 {
    ydir = -1
  }

  var x = x1, ywidth = y2 - y1, denom = x2 - x1
  while x <= x2 {
    var num = x - x1, ynext = y1
    if num == denom {
      ynext = y2
    } else {
      ynext = ynext + intdiv(ywidth * num, denom)
    }
    while y != ynext {
      noop x, y
      y = y + ydir
    }
    if y == ynext {
      noop x, y
    }
    x = x + 1
  }
}

proc draw {
  var i = 0
  while i < 100 {
    line 0, 0, i, 70
    i = i + 1
  }
  var j = 69
  while j >= 0 {
    line 0, 0, i, j
    j = j - 1
  }
}

draw`, nil)
	}
}
