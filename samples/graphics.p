proc fill {
  var i; i = 0;
  while i < 769 {
    line 0, i, 100, i
    i = i + 1
  }
}

proc design x, y {
  line x+0, y+0, x+100, y+100
  line x+0, y+100, x+100, y+0
  line x+0, y+0, x+0, y+100
  line x+0, y+0, x+100, y+0
  line x+100, y+0, x+100, y+100
  line x+0, y+100, x+100, y+100
  line x+0, y+0, x+50, y+100
  line x+0, y+0, x+100, y+50
}

export fill, design
