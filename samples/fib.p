#!/usr/bin/env pants2

func fib(x) {
  if x == 0 or x == 1 {
    return x
  }
  return fib(x-1) + fib(x-2)
}

println "what fibonacci number do you want?"
print "> "
var num
num = number(input())
println fib(num)
