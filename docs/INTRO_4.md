## Loop functions

Python has (with combination of iterators and range function) a powerfull and versatile __for loop statement__. 
It can iterate over a collection of items, loop over numbers, ranges, etc.

```python
for x in range(5):
    print(x)

primes = [2, 3, 5, 7]
for prime in primes:
    print(prime)
```

Rye takes a different approach here. Because loop constructs are __again just functions__ in Rye, you can have many of them, 
you can make our own, they can be defined in libraries which you can load if you need them.

Rye has less verbose equivavlents of the code above, which we will meet once we know __op-words__.

```factor
loop 5 { .print }

primes: { 2 3 5 7 }
for primes { .print }
```

But it also has more python / rebol like equivalents: 


```factor
repeat 'i 5 { print i }

primes: { 2 3 5 7 }
for-each 'prime primes { print prime }
```

Again, _repeat_ and _for-each_ are functions, both take 3 arguments. A word to set in each iteration, a value to iterate over or to and a block of code. 

> lit-word is another type of word. It means it doesn't evalueate to the value the word is linked to but to literal word itself.
> 'i and 'prime are lit-words above.

## Map, Reduce, Filter

You probably heard of these three. They are 3 popular higher order functions in _funcional programming_ languages. Semi functional languages have them also,
like Javascript and Python. Python is not particularly strong on FP side, Javascript is a little better. Let's look at map ...

These are functions in javascript and python too. That's why they are a little more limited.

Map _maps_ a list of items to new list of items given some (usually anonymous) function. 

Map function is passed an anonymous function.

```javascript
const items = [ 1, 2, 3, 4, 5 ];
const doubled = items.map(functionx(x) { return x * 2 });
const lower = items.filter(functionx(x) { return x < 4 });
```

Newer Javascript has a shorthand syntax for anonymous functions, so it can be written like this:
```javascript
const doubled = items.map(x => x * 2);
const lower = items.filter(x => x < 4);
```
Python also has map __function__, and a concept of lambda function, which is a small anonymous function that can only have one expression.

```python
items = [1, 2, 3, 4, 5]
doubled = list(map(lambda x: x * 2, items))
lower = list(filter(lambda x: x < 4, items))
```

Rye standard library has map-each (and map) too

```factor
items: { 1 2 3 4 5 }
doubled: map-each 'x items { x * 2 }  // this is normal block, not limited to 1 expression like in python
lower: filter-each 'x items { x > 2 }
```
Map each can take an anonymous function as third argument like the javascript or python above, but it can also take a block of code, which is 
less verbose and more lighterweight.

When we get to __op-words__ and "injected blocks", we will also be able to use map function

```factor
items: { 1 2 3 4 5 }
doubled: map items { * 2 }
lower: filter items { * 2 }
```

## User functions

This is how you define your (user) functions in Python:

```python
def greet(name):
    print("Hello, " + name + "!")

# and call it
greet("Jane")
```
And not so differently in Rye:

```factor
greet: fn { name } {
    print "Hello, " + name + "!"
}

# and call it
greet "Jane"
```

There are some differences behind the scenes though
* as always, greet set-word sets word _greet_ to the result of the expression on the right
* __fn__ is a function, that creates a function, it accepts 2 arguments, two block
    * first block is a argument list (a spec, it can include more than just arguments)
    * second block is code
    
And yet again. __fn__ is just a function, we could and do have many of those. For example _fnc_
that also accepts a context in which a function runs, a _closure_, and library specific ones.

> In next page we leave the planet of Rebol like Rye ...

_Next page soon_
