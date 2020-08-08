[&lt; Previous chapter](./INTRO_1.md)

## More about native functions

With all we know so far let's translate few Python snippets:

```python
name = "Jim"
print(name)
```

In rye this becomes code below. Notice the set-word instead of equals and lack of parenthesis after function call.

```factor
name: "Jim"
print name
```

Here we make a list, join items of it with a separator and print the result.

```python
names = [ "Jim", "Jane" ]
print(", ".join(names))
```

In rye, we make a _block_ (notice the lack of commas) then we join and print the result.

```factor
names: { "Jim" "Jane" }
print join names ", "
```

If you were a lisper, second line would look like this:

```factor
(print (join names ", "))
```

If you were in ptyhon or javascript it would be written like this:

```factor
print(join(names, ", "))
```
As we see Rye doesn't use parenhesis for function calls. Each word that is a function evaluates and "consumes" as much values 
to the right as it needs. That's why in Rye all functions must accept a fixed numbers of arguments.

## 'bout those blocks again

We used an list above and called it a block. As a python list, rye block can hold different types of values. Any types of values:

```python
some_list = [ "Jim", "Jane", 33, 12.5 ]
```

```factor
some-block: { "Jim" "Jane" 33 12.5 }
```

But remember, rye has 29 types of values, one of them are words.

```factor
a-block: { print join names ", " }
```
So what good is this now? ... We have this __do__ function that does a block

```factor
do a-block
Jim, Jane
```


