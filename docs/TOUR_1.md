<b><a href="./TOUR_0.html">A fantastic cereal</a> > Rye values and assignment</b>

# Rye values and assignment

## Rye value types

Rye knows many types of values. From numbers, strings, emails, URI-s, words, special words types, blocks of values ...

```rye
33
"Hello word"
jim@example.com
https://example.com
some-word
'print
data:
{ "Jane" "Jim" }
```

## Rye code = Rye values

Rye code is a block of Rye values. A block is also a Rye value.

```rye
print 11 + 22
print "Hello world"
data: { "Jane" "Jim" }
print first data
// prints:
// 33
// Hello world
// Jane
```

```rye
do {
  print 11 + 22
  print "Hello world"
  data: { "Jane" "Jim" }
  print first data
}
// prints the same as above
```
## Assignment

Words can be linked to Rye values. Set-words are used to create the assignment.

```rye
a-set-word: "takes value from the expression on the right"
name: "Jane"
age: 33
print name
new-age: print age + 1
// prints:
// Jane
// 34
```

<a href="./TOUR_2.html" class="foot next">Next</a>

### BONUS: Inline set-words

Everything in Rye is also an expression, returns a value. Setwords also return the assigned value, so they can be used inline.

```rye
prn "All:"
print all-fruits: 100 + apples: 12 + 21
prn { "Apples:" apples }
// prints:
// All: 133
// Apples: 33
```


### BONUS: Left and right set-words

They will make more sense later, but Rye also has left leaning set-words.

```rye
"Jim" :name
12 + 21 :apples + 100 :all-fruits
```
