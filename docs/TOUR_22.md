<p><b><a href="./TOUR_0.html">Meet Rye</a> &gt; Failures and errors</b></p>

# Check, fix and tidy

## Signaling intent

Rye language is big on signaling **intent**. Foreach can do anyting and more that you can do with map, filter and reduce. But the use of those three
has a benefit of signaling intent. Also a GOTO could do anything and more you can do with foreach, if, switch and even functions and you know the phrase ...

Rye has so far 3 main failure handling patterns so far: check, fix and tidy. We will see the basic functions below, but they are much more usefull in the form of **returning or skipping functions**.

## Check

if first argument:

* **is failure**: wraps failure in a failure constructed from second argument and returns that (re-fail).
* **isn't failure**: returns the argument

Check is usefull for checking if failure happened and **failure translation** that I mentioned in the <a href="./TOUR_21.html">first page</a>. It doesn't only translate (re-fail)
the failure, but it also wraps the parent failure, so you get the whole failure thread at the end. From the top level to the lowest.

	user profile could not be read > could not read user data file > missing file ./user-data.json

It can accept a **string** (failure message), an **integer** (failure status code), an **lit-word** (failure key), or a **block** (combination of above), to create a new failure value.

```rye
read-all %mydata.json |check { 404 "couldn't read the file" }
// returns:
//  a string of a file OR
//  a 404 failure wrapped around the failure of reading the file
```

## Fix

if first argument:

* **is failure**: does a block (second argument) and returns the result of evaluation to provide an alternative value
* **isn't failure**: returns the argument

```rye
get http://example.com/2134/username |fix { "Annonymous" } |print
// prints:
//  <username returned from the get request> OR
//  Annonymous

1 / 0 |fix { 50 } |print
// prints: 50
```


## Tidy

if first argument:

* **is failure**: does a block (to tidy after failure) and returns the failure
* **isn't failure**: returns the argument

```rye
get http://example.com/2134/username |tidy { "These was an error: " } |print
// prints:
//  <username returned from the get request> OR
//  These was an error:
//  <error structure>
```

## Returning and skipping functions

The patterns described above will be much more usefull in their special forms, which I will describe on the next <a href="./TOUR_23.html">two</a> <a href="./TOUR_24.html">pages</a>.