[&lt; Previous chapter](./INTRO_2.md)

## A man can't live of ~~bread~~functions alone

Ok, so we were calling few functions now, but we forgot about all the other basic stuff, like the __if__ statement, the __loops__.

First of all, rye doesn't have any statements, everything is an expression, everything returns something.

```python
if 10 < 100:
   print("10 is less than 100")
```
_Do_ you remember the __do__ function? I accepts the block of code as a first argument.

```factor
do {
   print "10 is less than 100"
}
```
Could there be a similar function, but it would _do_ the code (now second argument) only if the first argument
would be true? Yes ... 

```factor
if 10 < 100 {
   print "10 is less than 100"
}
```



_more later .. currently we have greater? lesser? instead of < > in rye implementation, so I will first add these before we continue_
