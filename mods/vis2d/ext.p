proc fill x1, y1, x2, y2 {
  if x2 < x1 {
    var x2t, y2t
    x2t = x2; x2 = x1; x1 = x2t
    y2t = y2; y2 = y1; y1 = y2t
  }

  var i; i = y1;
  while i < y2 {
    var j; j = x1;
    while j < x2 {
      pixel j, i
      j = j + 1
    }
    i = i + 1
  }
}

func intdiv(x, y) {
  var m
  m = x % y
  m = x - m
  return m / y
}

proc line x1, y1, x2, y2 {
  if x2 < x1 {
    var x2t, y2t
    x2t = x2; x2 = x1; x1 = x2t
    y2t = y2; y2 = y1; y1 = y2t
  }

  var y, ydir
  y = y1
  ydir = 1
  if y2 < y1 {
    ydir = -1
  }

  var x, ywidth, denom
  x = x1
  ywidth = y2 - y1
  denom = x2 - x1
  while x <= x2 {
    var num, ynext
    num = x - x1
    ynext = y1
    if num == denom {
      ynext = y2
    } else {
      ynext = ynext + intdiv(ywidth * num, denom)
    }
    while y != ynext {
      pixel x, y
      y = y + ydir
    }
    if y == ynext {
      pixel x, y
    }
    x = x + 1
  }
}

export line, fill, intdiv
