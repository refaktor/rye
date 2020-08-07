# Introducing Rye to Python programmers

Python is a clear, understood language that many people know. So I will try using Python examples to introduce you to the Rye language.

But Python and Rye are quite different. All Rye's core ideas are taken from __Rebol__. Words, blocks, code is data, etc.. 
It also takes some aspirations from __Factor__ (a stack based language) and some from your ordinary __linux shell__ (pipes). 

## Theory be damned!

>When I say Rye below, I most of the time mean the Do dialect of Rye ... more on that later

Let me tell you a little about the basic principles in Rye, so the code forward will make more sense. 

It's quite simple, there are no keywords, statements or special forms.

### Nothin' but values

Rye language is currently composed of 29 different types of __values__. Values can be literal values (numbers, strings, dates), words, blocks, etc.

You could say that Rye code is nothing but values and it would be true. 

```factor
1
2
"Jim"
```
Python has 13 data types [[py1]](https://www.w3schools.com/python/python_datatypes.asp), but it has a lot of other "things".

### Block of Values

A more interesting type of value is a __block__. Because block holds other values inside. Value of values :)

```factor
{ 1 2 "Jim" }
{ { "Jim" 33 } { "Jane" 35 } }
```
Block is like a list in Python.

### Words

Probably the most interesting type of value is a __word__. Words can be linked to any other value. Like we saw above, 
to numbers, strings, blocks, other words, etc. There are multiple types of words, actually. First two we will meet
are __set-words__ and just ordinary __words__.

```factor
age: 101
description: "wise"
```
The colon on the right identifies a __set-word__. A set word in Rye evaluates expression on the right and
sets the value of it to that word.

```factor
print age
101
print description
wise
```
Ordinary __words__ like _age_ and _description_ just return the values they are linked to. 

So, words are similar to variables in Python. And what is __print__ then?

### Native functions

__print__ above looks a lot like __age__ and __description__. It looks like a word and it is a word. But this word is 
linked to another type of value, a __native function__. 

Native behind _print_ accepts 1 argument, so rye evaluates expression on the right to get that value. Then the native 
prints it.

Better late than never, now would be a good time to do:

```factor
print "Hello world!"
Hello world
```

_Next chapter soon_ 
