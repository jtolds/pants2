#!/usr/bin/env pants2

var ans
ans = random(1, 101)
println "I'm thinking of a number between 1 and 100. What is it?"
loop {
  print "> "
  var guess
  guess = input()
  if guess == "" { break }
  guess = number(guess)
  if guess < ans {
    println "guess is low"
    next
  }
  if guess > ans {
    println "guess is high"
    next
  }
  println "you got it!"
  break
}
println
