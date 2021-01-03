<b><a href="./TOUR_0.html">A fantastic cereal</a> > User functions</b>

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


## Currying

This is experimental, but Rye has a form of partial application.

```rye
add5: add _ 5
subtract-from-10: subtract 10 _

print add5 10
// prints: 15

print subtract-from-10 3
// prints: 7

```

