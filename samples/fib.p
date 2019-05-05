#!/usr/bin/env pants2

func fib(x) {
  if x == 0 or x == 1 {
    return x
  }
  return fib(x-1) + fib(x-2)
}

print "what fibonacci number do you want?"
print "> "
var num = number(input())
print fib(num)

loop {
  sleep 10
}
