## Loop functions

Python has a powerfull and versatile __for loop statement__ (with combination of iterators and range function). 
It can iterate over a collection of items and loop over nubers (ranges).

```python
for x in range(5):
    print(x)

primes = [2, 3, 5, 7]
for prime in primes:
    print(prime)
```

Rye takes a different approach here. Because for loop is __again just a function__, you can have many of them, you can make our own, they can be defined 
in libraries which you can load if you need them.

Rye has less verbose equivavlents of the code above, which I will go through after you read about __op-words__ and __pipe-words__

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

Again, repeat and for-each are functions, both take 3 arguments.

## The function creating function

Ok, you get it. All these things in rye are functions, __if, either, for-each, repeat__ and yet unseen switch, case, match, loop ...

