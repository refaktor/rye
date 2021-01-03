<b><a href="./TOUR_0.html">A fantastic cereal</a> > Loop and for</b>

# Loop and for

## Just functions again

Looping constructs in Rye are also just functions. Rye has many of them, and you can create your own.

## Simple loop

Sometimes you just need to loop N times

```rye
loop 2 { prn "bye" }
// prints: bye bye
```
If you need an index of repetition, it's injected into a block.

```rye
loop 3 { :i prn i + 1 }
// prints: 1 2 3 
```

The use of left set-word might seem a little odd, but you will see the benefits later.

## For function

For functions iterates on a collection of values.

```rye
names: { "Jim" "Jane" "Anne" }
for names { :name print "Hi " + name + "!" }
// prints:
// Hi Jim!
// Hi Jane!
// Ji Anne!

for range 1 5 { :i print l-pad i i "0" }
// prints:
// 1
// 02
// 003
// 0004
// 00005
```

<a class="foot" href="./TOUR_4.html" class="next">Next</a>


### BONUS: Use of pipe and op-words

We will come to them later, but we would write the examples above like this:

```rye
loop 3 { + 1 |prn }
// prints: 1 2 3

for names { .embed "Hi {{}}!" |print }
// prints:
// Hi Jim
// Hi Jane
// Ji Anne
```
