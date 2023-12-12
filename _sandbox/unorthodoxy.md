# Unorthodox sides of Rye

_Being unorthodox offers no benefit by itself._ 

_If you aren't creating anything really new, just use the old._

Rye is a programming language, no different than <Insert your own>. Here I am trying to document sides of Rye that can be somewhat different, because the things that aren't, you already understand.
Many of these specifics come from Rye's inspirators, Rebol, Factor, Shell and Go. 

## Newlines and spaces

Whole Rye program could be written in one line. There are no statement separators (there are also no statements, just expressions btw).

Each Rye token must be separated by a space.

Rye knows a token "," called expression guard, which is optional, but can help you explicitly safeguard the limits of expressions.

----

_Unordered_

## Coding conventions

* ? at the end of the word stands for get- and is usually used when a word is more of a property than a verb. Like: length? , color?
* functions that return boolean values often use is-xxxx. Like: is-integer , is-positive

## Op-words and Pipe-words

## Returning words

## Failures and Errors

## Hierarcies of Contexts

## Scopes (Searching for context)
