# Unorthodox sides of Rye

_Being unorthodox offers no benefit by itself._ 

_If you aren't creating anything really new, just use the old._

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
