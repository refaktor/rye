[&lt; Previous chapter](./INTRO_1.md)

## More about native functions

With all we know so far let's translate few Python snippets:

```python
name = "Jim"
print(name)
```
In rye this becomes code below. Notice the set-word instead of equals and lack of parenthesis.

```factor
name: "Jim"
print name
```
Here we make a list, join them with separator and print the result.

```python
names = [ "Jim", "Jane" ]
print(", ".join(names))
```
In rye, notice another lack, the lack of commas.

```factor
names: { "Jim" "Jane" }
print join names ","
```
You could understand the second line as:

```factor
(print (join names ","))
```
In Rye, all functions must accept a fixed number of arguments.
