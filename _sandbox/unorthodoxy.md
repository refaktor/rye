# Unorthodox sides of Rye

_Being unorthodox offers no benefit by itself._ 

_If you aren't creating anything really new, just use the existing._

_**work in progress document**_

Rye is a programming language, not that different than _insert-your-own_. Here, I will try to document sides of Rye that are somewhat different. Because the things that aren't, you already understand.
Some of these specifics come from Rye's inspirators, Rebol, Factor, Shell and Go. Some are unique. 

## Newlines and spaces

Whole Rye program could be written in one line. There is no concept of lines, no line separators.
```red
print 1 + 100 + 1000 prns "Hello" if 1 { loop 3 { print "World" } } print "Bye"
```

Each Rye token must be separated by a space. This also means that parenthesis or operators must always be one space removed from it's neighbouring tokens. 
```red
print 123+234 ; loader error
if 1 {print "OK"} ; loader error
```

Rye knows a token "," called expression guards. They are completely **optional**. You can use them so explicitly mark expression boundaries, if you want to.
```red
print 1 + 100 + 1000 , prns "Hello" , if 1 { loop 3 { print "World" } } , print "Bye"
```

## Conventions around nouns, verbs and adjectives ...

Words in Rye can be bound to Rye values. We try to implement these conventions ... we aren't always sucsessfull as we also want to be succint and lean on the
some general conventions everyone is already used to.

* **Nouns** are usualy bound to concrete values (person, age, email)
* **Verbs** are usually bound to functions (send, print, inspect)
* **Adjectives** usually return boolean results (is-hot, is-cold, is-red) and start with "is-"
* **Nouns** starting with "is-" also return boolean result (is-integer, is-email, is-zero)
* **Nouns** ending with **?** stand for **get-**. Usually it's used when a **noun** is a property (length?, color?, age?)
* **Nouns** that are types or kinds in the language are constructors (fn, context, dict, list, table)

More conventions
  
* If a function performs a same task as another, but in a different way or with a different (number of) arguments it can be defined with **\variation** at the end (print\val, load\csv, map\pos)
* If a function ends with **\\**, this means **"more"**, usually means a this variation accepts additonal argument compared to base function (ls\, produce\)
* If a functions changes values **in-place** it has ! at the end (inc! , append! , unique!)

Two types of Rye functions exist. General functions and generic functions. 

## All the word types!?

Words, set-words, lset-words, get-words, tag-words, ... why ...

## Assignment, right, left, inline ...

```red
name: "Jim"
20 :age
....
```

## Functions as Words, Op-words, Pipe-words and an odd Star

Rye words that are bound to functions (user or builtin) can be called as words in normal reverse polish notation.
```red
print "Hello"
print capitalize "jane"
print concat "Am" "mo"
```
The same function can be called as an op-word (with a . infront), in this case it takes it's first argument from the left.
```red
"Hello" .print
"jane" .capitalize .print
print "Am" .concat "mo"
```

Difference is in priority. Opwords take the first argument on the left they can find, pipewords evaluate all expressions on the left and take the result as argument.
```red
"Hello" |print
"jane" |capitalize |print
"Am" .concat "mo" .print ; will print mo
"Am" .concat "mo" |print 
```

Operators are implicitly op-words, without the need to add them .. Their spelling is _+ _*. They can be called as pipe words too.
```red
1 + 2
23 + 33 * 32 / 2
23 + 33 * 32 |/ 2
_+ 23 43
```

## Returning words

Rye doesn't use return statement. It always returns the last value (akin to rebol, lisp, ...). You can use return function to return before that. But also other functions can return or conditionaly return. Because we want them
to be obviously visible, convention is that they should have ^ at the begining. Such words are ^if, and some of the failure handling functions ^fix ^check 

```red
check: fn { name } { ^if name = "foe" { "go away" } "hello" }

check "friend" ; hello
check "foe"    ; go away
```

## Failures and Errors

Programmer's errors (bugs) and failures are not the same. Failure is a (not unexpected) result of a function that couldn't do (return) what it is expected to do. 
Failure at this point is information and can be handeled with Rye functions. If it's not handeled it becomes  a programmers error (a bug). Failures can be handeled, bugs gave to be corrected in code.

## Multimethods

Generic methods that dispatch on the kind of first argument.

## N datatypes, M syntax types

Rye like Rebol has many datatypes and many of those are syntax types, meaning you can enter them directly through code with no conversion. This helps at declarative code because more specific types let you create richer more information rich structures.

## Blocks are your bread and butter

Blocks are like lists or arrays, but also all Rye code lives in blocks and you structure your code in blocks.

## Context is everything

Context is another datatypes of Rye that is heavily used by the language itself. All code is executed in a context. 

## Table datatype

Some languages claim that they are high level languages. I claim for Rye, that it's even higher (than usual) level language. For one, if language want's to be higher level it should also have higher level structures. Higher 
usually means more towards the human. So I believe structures should also be speaking more about human's view on information, than some computer science concept ...

## Validation dialect

A big part of (safe) code goes into validating values. I believe this part doesn't need to be intermingled with so called "business logic", but can be a separate, better visible, declaratively defined part. So a core Rye includes validation dialect.

## Conversion dialect

A big part of code is just converting A's in B's and A1's into A2's. Since A, B, A1 and A2 are constants the conversion between them could be constant and could always be reused and cleaned from the regular code.

## Stance on dialects and macros

_work-in-progress-text_

I am not expert on Lisps, not even Rebol, I am just speaking my mind. You are welcome to tell me I am wrong (janko.itm at gmail).

Rebol has dialects, Lisps have macros. Each of them have benefits, but they also incur a cost, as they are an exception(1).
An optimisation in syntax, elegance, reduction in code, but still an exception. Exception you have to **know** **about**, **notice**, **learn**, **understand** (3) ...

In lisps, **macros** can be casually mixed with non-macro words (functions). Well, they by default usually are, since as macros alongside functions form many foundational constructs. *if* (2), *defn*, *loop* are usually macros.
Macros can change the rules of the game. So you have this maximaly uniform language where code and data are uniformly represented as lists, but any token inside this language could change the rules. That's why many Lisp authors 
warn about macro overuse.

```
Guy Stelle (of Scheme) said: "Macros are a powerful tool, yet they should be used sparingly. Overuse of macros makes programs hard to read, hard to debug, and hard to maintain."
Doug Hoyte (of Let over Lambda): "Macros should be invisible to the user or be more beautiful than the alternative."
```

My _naive idiot's_ opinion is, that macros in lisps (as these powerful but potentialy dangerous constructs) maybe have suboptimal distribution/visibility.

I dare not to really compare Rebol to the great family of languages as Lisps are (now I made both Rebolers and Lispers mad). But Rebol's approach avoids some of the troubles lisps have. Some say
Rebol is like a Lisp without parenthesis, but we don't care about that here. In Rebol, blocks (akin to lists) do not auto-evaluate, a significant departure from Lisp's default behavior. This miniscule detail 
changes a lot. You don't need any special forms, because functions can accept
blocks of code directly (since they don't get evaluated by runtime) ... so *if*, *defn*, *loop*, etc are just functions. Functions like any other functions. So Rebol doesn't have or _need_ macros. 
But it has **dialects**. It has really interesting approach to dialects, with it's parse _dialect_ (I know :P). Dialects are like special interpreters for Rebol tokens. They have better separation (a benefint in clarity, a cost in reuse)
than macros, they always exist in their own blocks. They aren't sprinkeled aroung your regular Rebol code, but again ... the more you use them the more full of specail cases (syntaxes, evaluation rules) your code becomes.

So my again _naive idiot's_ opinion is, dialects are NET benefit in only one case: 

For **big, "famous"** cases. Cases you will **know about**, will **notice**, **learn** about them, and invest to **understand** (3) them. Rye has a validation dialect in it's core. If you program in Rye, you know about
validation dialect as you know about for loop in Python. But they don't have to be too easy to make, main investment in dialects is in using them, so if making them takes a little Go, I see no problem for now.

I think Reboler's view is also that dialects are good solution for declaring specialised, local, problem specific solutions. But I think regular Rye, with the flexibility of custom **contexts** can better solve these. I also 
think that the core language could be flexible enough for specific usages like GUI (3) where Rebol employs the famous Rebol VID dialect.
 
(1) not exception as an error, but exception to the normal operation, evaluation
(2) well *if* or eq. is not just a macro but a _special form_ because it requires special evaluation rules that aren't compatible with lisps evaluation. Lisp by default immediately evaluates lists. 
(3) initial tests with regular Rye code using GTK gave nice results

## Is Rye Object Oriented

No. I believe that separation between functionality and state is usualy a better approach. You can use contexts as objects, you could also extend them, put them in hierarchies and in general do anything you would 
when using a prototype-based object oriented programming language like Lua, Self or Javascript.

## Focus, orientation, vocabulary

Most of the early computer programming languages were written with focus tt the computer. If you look at it's words and structure they are computer related words, data structures, functions.

Another group of languges seems aimed and focused on the model of computation at their grand idea. The naming, structures, functions are those which elevate the model. Lisp is all about lists and functions, many functional languages stem from mathematics.

I am not saying that is bad and we are better. No. Just that Rye is consciously focused on user (the human) and on problem description.





