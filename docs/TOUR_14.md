<b><a href="./TOUR_0.html">Some Practical Rye</a> > Pipe-words, op-words</b>

# Pipe-words, op-words

## From words to pipe-words

```rye
print add 100 inc 10
// prints: 111

// is the same as
10 |inc |add 100 |print
```
Each native or user function can be used as pipe-word.

## Op-words vs Pipe-words

```rye
add 10 .print 20 .print |print
// prints:
// 10
// 20
// 30
```
