<b><a href="./TOUR_0.html">Meet Rye</a> > User functions</b>

# User functions

## Define a function

```rye
double: fn { a } { a + a }

print double "Hey"
// prints: HeyHey
```

surprise, surprise ... fn is also a (native) function!

## Does and pipe

If function doesn't accept any arguments we can define it with does. If one with pipe.

```rye
inc: fn { a } { a + 1 }
hey: fn { } { print "Hey" }

// you can use
inc: pipe { + 1 }
hey: does { print "Hey" }

hey hey
// prints:
// Hey
// Hey

print inc 10
// prints: 11

10 |inc |print
// prints: 11
```

Function is a first class Rye value. We usually call them by invoking the word the got assigned to.

```rye
apply-1: fn { val mod } { mod val }

apply-1 10 pipe { + 1 } |print
// prints 11
```

If you invoke a word that holds a function, the function gets called. If you want to get a function itself you can use get-word (?get-word). 
```rye
apply-1: fn { val mod } { mod val }

increment: pipe { + 1 }

apply-1 10 ?increment |print
// prints 11

```

## Pure functions

Pure functions are functions that have no side effects and referentially transparent. You can define your own pure functions and they must
call just other pure functions or natives.

Pure functions only have access to pure context, so for them any unpure words are simply undefined.

```rye
add3: pfn { a b c } { a + b + c }

add3 1 2 3 |print
// prints: 6

non-pure: pfn { a b c } { print a + b + c }
non-pure 1 2 3
Error: Error: word not found: print 
At location:
{ <-here-> print a ._+ b }
```


## Currying

This is somewhat experimental, but Rye has a form of partial application. It can partially evaluate on any argument.

```rye
add5: add _ 5
subtract-from-10: subtract 10 _

print add5 10
// prints: 15

print subtract-from-10 3
// prints: 7

add10: 10 + _
print add10 5
// prints 15

{ 10 100 1000 } |map 1 + _ |for { .print }
// prints:
// 11
// 101
// 1001

db: open sqlite://test.db
myquery: query db _
myquery { select * from pals } |print
// prints:
// | id | name |
// | 1  | Jane |

```

