<b><a href="./TOUR_0.html">A fantastic cereal</a> > Map, filter and HOF-s</b>

# HOF-s

In functional programming languages, functions are first class values. Functions that accept them as arguments are called higher-order-functions (HOF-s).

Most known HOF-s are map, filter and reduce. Rye functions are also first class values, but in rye code blocks are also first class.

## Map

Map a block of values to a block of different values.

```rye
nums: { 1 2 3 }

map nums { 33 }
// returns: { 33 33 33 }

map nums { + 30 }
// returns: { 31 32 33 }

map nums { :x all { x > 1  x < 3 } }
// returns: { 0 1 0 }

strs: { "one" "two" "three" }
{ 3 1 2 3 } |map { - 1 |<- strs |to-upper } |for { .print }
// prints:
// THREE
// ONE
// TWO
// THREE
```

## Filter 

Filter returns a block of values where the block of code was Truthy.

```rye
nums: { 1 2 3 }

map nums { 33 }
// returns: { 1 2 3 }

map nums { < 3 }
// returns: { 1 2 }

map nums { :x all { x > 1  x < 3 } }
// returns: { 1 }

strs: { "one" "two" "three" }
{ 3 1 2 3 } |filter { > 1 } |map { <-- strs } |for { .print }
// prints:
// three
// two
// three
```

## Use with natives and curry

Instead of a block of code hofs currently also accept native functions and curried native functions

```rye
nums: { 1 2 3 }

map nums ?inc
// returns 2 3 4

maps nums 30 + _
// returns 31 32 33

nums |filter 1 > _ |map 10 + _ |for { .prn }
// prints: 12 13
```

Support for user functions and curried user functions still needs to be implemented. You will be able to read more about curried functions on additional page.

## More HOF-s

There will be more HOF-like functions. We already also have seek and purge. Reduce and sumize are waiting for another potential language feature (packs) I am thinking of which would make
their code more elegant.

```rye
{ 1 2 3 } .map fn { x } { x < 3 }
// returns: { 1 2 }
```
