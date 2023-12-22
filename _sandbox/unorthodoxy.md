# Unorthodox sides of Rye

_Being unorthodox offers no benefit by itself._ 

_If you aren't creating anything really new, just use the old._

_**work in progress document**_

Rye is a programming language, no different than _insert-your-own_. Here I will try to document sides of Rye that are somewhat different. Because the things that aren't, you already understand.
Many of these specifics come from Rye's inspirators, Rebol, Factor, Shell and Go. Some are unique. 

## Newlines and spaces

Whole Rye program could be written in one line. There is no concept of lines, no line separators.

Each Rye token must be separated by a space. This also means that parenthesis or operators must always be one space removed from it's neighbouring tokens. 

Rye knows a token "," called expression guard, which is optional. It can help you explicitly safeguard the limits of expressions or visually signify them.

## Conventions around nouns, verbs and adjectives ...

Words in Rye can be bound to Rye values. We try to implement these conventions ... we aren't always sucsessfull as we also want to be succint and lean on the
some general conventions everyone is already used to.

* **Nouns** are usualy bound to concrete values (person, age, email)
* **Verbs** are usually bound to functions (send, print, inspect)
* **Adjectives** usually return boolean results (is-hot, is-cold, is-red) and start with "is-"
* **Nouns** starting with "is-" also return boolean result (is-integer, is-email, is-zero)
* **Nouns** ending with **?** stand for **get-**. Usually it's used when a **noun** is a property (length?, color?, age?)
* **Nouns** that are types or kinds in the language are constructors (fn, context, dict, list, spreadsheet)

----

## Op-words and Pipe-words

## Returning words

## Failures and Errors

## Hierarcies of Contexts

## Scopes (Searching for context)

## Multimethods

## N datatypes, M syntax types

## Spreadsheet datatype

## Validation dialect

## Conversion dialect

## Stance on dialects and macros

_~work-in-progress-text~_

I am not expert on Lisps, not even Rebol, I am just speaking my mind. You are welcome to tell me I am wrong (janko.itm at gmail).

Rebol has dialects, Lisps have macros. Each of them have benefits, but they also incur a cost, as they are an exception(1).
An optimisation in syntax, elegance, reduction in code, but still an exception. Exception you have to **know** **about**, **notice**, **learn**, **understand** (3) ...

In lisps, **macros** can be casually mixed with non-macro words (functions). Well, they by default usually are, since some of the basic construct of a language are macros. *if* (2), *defn*, *loop* are usually macros.
Macros can change the rules of the game. So you have this maximaly uniform language _where everything is a list_, but any token inside this language could change the rules. That's why many Lisp authors warn about macro overuse.
Guy Stelle (of Scheme) said: "Macros are a powerful tool, yet they should be used sparingly. Overuse of macros makes programs hard to read, hard to debug, and hard to maintain."
Doug Hoyte (of Let over Lambda): "Macros should be invisible to the user or be more beautiful than the alternative."
My _naive idiot's_ opinion is, that macros in lisps (as these powerful but potentialy dangerous constructs) maybe have suboptimal distribution/visibility.

I dare not to really compare Rebol to the great family of languages as Lisps are (now I made both Rebolers and Lispers mad). But Rebol's approach avoids some of the "troubles?" lisps have. Some say
Rebol is like a Lisp without parenthesis, but we don't care about that here. Rebol's blocks (lists) DON'T evaluate by default. This miniscule detail changes a lot. You don't need any special forms, because functions can accept
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


