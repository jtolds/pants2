#!/usr/bin/env pants2

var ans = random(1, 101)
print "I'm thinking of a number between 1 and 100. What is it?"
loop {
  print "> "
  var guess = input()
  if guess == "" { break }
  guess = number(guess)
  if guess < ans {
    print "guess is low"
    next
  }
  if guess > ans {
    print "guess is high"
    next
  }
  print "you got it!"
  break
}

loop {
  sleep 10
}
