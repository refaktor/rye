[&lt; Previous chapter](./INTRO_2.md)

## Expressions all the way down

So we tried calling few functions so far, but we forgot about all the other basic stuff, like the __if__ statement, the __loops__.

First of all, rye doesn't have any statements, everything is an expression, __everything returns something__. 

It also doesn't need explicit __return__ statements, the result of last expression in the block is returned. A little weird piece
of code:

```factor
print 100 .add do { print 1 print 11 }
1
11
111
```
The last expression in a block is `print 11`, print returns the value it prints, so the block returns 2. You can figure out the rest.

## Function calls all the way down

In most languages __if__ is a statement, or a __special form__, feature of the language. Even in Lisps, if is a macro.

```python
if 10 < 100:
   print("10 is less than 100")
```

Rye doesn't evaluate blocks by default, so they can be ordinary arguments. _Do_ you remember the __do__ function? I accepts the
block of code as a first argument.

```factor
do {
   print "10 is less than 100"
}
```
Could there be a similar function, but it would _do_ the code (now second argument) only if the first argument
would be true? Of course ... 

```factor
if 10 < 100 {
   print "10 is less than 100"
}
```

If __if__ is just a function, it means you can make your own if-like functions inside rye. 

## U either like it or U don't

The downside of this is that in Rye we can't have special forms like __if ... else ....__, because _we don't
have special forms at all_. 

```python
if name == "Jim":
   print("Hi Jim")
else:
   print("Door is locked")
```

So we have a function called __either__, that takes additional block as argument. 

```factor
either name = "Jim" {
   print "hi Jim"
} {
   print "Door is locked"
}
```

See, it's not so bad. But because everything is a function call it makes a language much more uniform,
simpler and malleable.

And I said, everything in Rye retuns something, so the example above would be better written as:

```factor
print either name = "Jim" { "hi Jim" } { "Door is locked" }
```

## About those turtles

Someone once said "Turtled all the way down". We've learned in this page that in Rye (do dialect) there is:

* Rye values all the way down
* Expressions all the way down
* Function calls all the way down

_Next page sooner or later_
