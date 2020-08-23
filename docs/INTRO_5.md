<a href="./INTRO_3.html" class="prev">Previous page</a>

## Rye and Rebol

_this page is work in progress_

There is more to the Rebol-s than that, but up until now, everything that I wrote about, could also be an intro to Rebol, Red or Ren-c languages.

The core concepts of all 4 languages come from Rebol. What we build on that core differs and here Rye strays on some new grounds, a little aspired
mostly from Factor and linux shell. The forward moving evaluation ...

## Meet op-words

Rye adds another type of value. An op-word (operator-word?), bacause it behaves most like operators would. Op word is identified by a dot on the left.

```rebol
print add 4 5
9
// with op word you can do
print 4 .add 5
9
```
you can use any word that links to a native or a user function as an opword

```rebol
// with op word you can do
{ 1 2 3 } .map { + 1 }
```

Operators like + - * > < are automatically recognised and used as op-words.

Op-words can take any positive number of arguments.

```rebol
print add inc 2 2
5
// with op word you can do
2 .inc .add 2
5
```

## And the pipe-word

What if we want to use print in example as op-word? 

```rebol
// with op word you can do
print 4 .add 5
```
We will see that it doesn't do what we want, it prints 5 instead of 9.

```rebol
4 .add 5 .print
5
```
Op-word takes the first value on the left that it can and proceed. So if you want all expressions on the left to evaluate and call a 
function result of that, you can use the pipe-word.


```rebol
// with op word you can do
4 .add 5 |print
9
```

Maybe as a little fun/weird example. We can have functions like _skip_. It accepts 2 arguments, a value and a block of code.
It executes a block of code and it returns the first argument.

```rebol
4 .prn .skip { prn "+" } |add 5 .print |print
4+5
9
```

 * 4 evaluates to 4
 * .prn takes 4, prints it and returns it
 * .skip taks this 4, executes the block, and returns it
 * |add takes this 4 and looks to the right for a secod argument
 * there is finds 5 but to is a .print that takes precedence, prints 5 and returns 5
 * add gets 5 as second argument, adds 4 and 5 together and returns 9
 * |print prints it


<a href="./INTRO_4.html" class="prev">Previous page</a> -
<a href="./INTRO_6.html" class="next">Next page</a>

> Next we will look at more language utilities, to support this left-to-right style
> The lset-words and a concepts so far called injected blocks
