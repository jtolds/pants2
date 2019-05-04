#!/usr/bin/env pants2

import "./mods/vis2d/ext.p"

proc box x, y {
  color "green"
  line x*100, y*100, x*100, (y+1)*100
  color "red"
  line x*100, y*100, (x+1)*100, y*100
  color "blue"
  line x*100, y*100, (x+1)*100, (y+1)*100
  color "blue"
  line x*100, (y+1)*100, (x+1)*100, y*100
  color "green"
  line x*100, (y+1)*100, (x+1)*100, (y+1)*100
  color "red"
  line (x+1)*100, y*100, (x+1)*100, (y+1)*100
}

proc boxes {
  drawoff
  var i = 0
  while i < 7 {
    var j = 0
    while j < 10 {
      box j, i
      color "white"
      locate j*100 + 6, i*100 + 3
      print i, j
      j = j + 1
    }
    i = i + 1
  }
  drawon
}

proc linedesign {
  drawoff
  var i = 0
  while i < 1000 {
    color "green"
    line 0, 0, i, 700
    i = i + 1
    color "red"
    line 0, 0, i, 700
    i = i + 1
    color "blue"
    line 0, 0, i, 700
    i = i + 1
    if i % 33 == 0 { drawon; drawoff }
  }
  drawon; drawoff
  var j = 699
  while j >= 0 {
    color "green"
    line 0, 0, i, j
    j = j - 1
    color "red"
    line 0, 0, i, j
    j = j - 1
    color "blue"
    line 0, 0, i, j
    j = j - 1
    if j % 33 == 0 { drawon; drawoff }
  }
}

boxes
linedesign
boxes

loop {}
