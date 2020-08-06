# Introducing Rye for Python programmers #1

Python is a clear, understood language that many people know. So I will try using Python examples to introduce some basic ideas in Rye.

But Python and Rye are quite different. All Rye's core ideas are taken from Rebol. Words, blocks, code is data, etc.. 
It also takes some aspirations from Factor (a stack based language) and just from using an ordinary linux shell (pipes). 

## Theory be damned!

>When I say Rye below, I mean the Do dialect of Rye ... more on that later

### Nothin' but values

Rye language is currently composed of 29 different types of __values__. Values can be literal values (numbers, strings, dates), words, blocks, etc.

You could say that Rye code is nothing but values and it would be true.

```factor
1
2
"Jim"
```

### Block of Values

A more interesting type of value is a __block__. Because block holds other values inside. Value of values :)

```factor
{ 1 2 "Jim" }
{ { "Jim" 33 } { "Jane" 35 } }
```

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
Ordinary __words__ like age and description just return the values they are linked to. 

So what is __print__ then?

>Next chapter soon 
