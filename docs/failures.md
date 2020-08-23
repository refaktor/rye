# Failures and Errors

_work in progress_

When I started writing Rye I had no specific ideas about Exception handling. But the classic _try catch_ model looked annoying to me.

I had a problem with it visually, the code without exceptions flowed, but if you added all exception handling that needed to be there (sometimes) the 
code became a mess. Try catch added code structure, where there logic wise wasn't one. And many times (or in many languages) exception handling looks to me
like a goto statement.

I wanted to make something that is visually and structurally not so obtrusive. I wanted something in-flow. I composed something based on that, we have yet to
see if it really works any better than the classic approach.

Everything below this point is just a current hypothesis, it may be wrong in parts or alltogether, but let's play with it and see.

## Types of exceptions

Let's try to start from zero. 

  * Some exceptions are a result of __programming bug__. It means the program should stop (as we don't know what will happen next and what should). Exception should
  becommunicated to the user and logged for programmer, to fix the bug. We don't catch and handle these exceptions, we fix them if we know for them.
    * static bugs
      * syntax errors  - the code is not loadable rye, example: "err: {123 asda }"
      * naming errors - words that aren't defined or are misnamed
      * structure errors - the code isn't structured as used words would need it to be: "loop 2 3" ... 3 should be a block of code
    * runtime bugs
      * __unhandeled__ value errors - the value or type of value is such that evaluation can't proceed: division w/ zero, wrong type
      * __unhandeled__ io errors
  * Runtime exceptions, that you predicted can happen, and you check for them or handle them after they happen, sometimes they can be used
    to controll logic (example 1.)
      * value failures (wrong type, division by zero, conversion failures, parsing failures, ...)
      * io failures (filesystem related like: disk out of space, insufficient priviliges, nonexistent path ; network errors ; ssl errors ...

## What should happen

  * When a __bug happends__ the program should notify the user, log the error, and stop execution in all cases except maybe in server environments, where you
  want continious running and only the process or procedure is stopped
  * When runtime failure happens it shoul be handeled, if not it's a bug so first applies
  
## The sto stages of runtime exceptions in rye

Value and IO exceptions start as __failures__. Failure to do the desired operation. If failure is not handeled or returned it becomes an program error and stops the execution.

## In what ways do we handle failures

  * we can return in to caller function
  * we can wrap it into a higher level description of error and return it
  * we can provide the alternative / default value instead of computed one
  * we can do some action (like cleanup, get alternative value ...)
  
## Some examples from python

```python
while True:
  try:
    x = int(input("Please enter a number: "))
    break
  except ValueError:
   print("Oops!  That was no valid number.  Try again...")
```


```rye
 while {
   input "Please enter a number:" 
     |to-int
     |fix-either 
       { print "This was not a valid number. Try again" }
       { .print-val "You entered number {#}." false }
 ```
  
