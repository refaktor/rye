# Failures and Errors

_work in progress_

_This document is a thought experiment in progress. I am trying to test the current idea I had about exceptions. I am fully aware, that it might 
be insufficient, not solution at all or worse than status quo that I am criticising (to see what really is the problem, or I think it is). I am testing it out by writing this doc_

## Exceptions are where beautifull code goes to die

When I started writing Rye I had no specific ideas about Exception handling. But the classic _try catch_ model looked annoying to me.

I had a problem with it visually, the code without exceptions flowed, but if you added all exception handling that _needed to be there_ (many times) the 
code became a mess. Try catch added code structure, where there logic wise wasn't one. And it's sort of GOTO-ish structure at that.

Try catch statements are often times not specific enough, and at the same time much too verbose. 

This is an example from python docs:

```python
try:
    f = open('myfile.txt')
    s = f.readline()
    i = int(s.strip())
except OSError as err:
    print("OS error: {0}".format(err))
except ValueError:
    print("Could not convert data to an integer.")
except:
    print("Unexpected error:", sys.exc_info()[0])
    raise
```

The problems I have with this:

  * visual: quite sequetial logic is now split in 3 code blocks with unclear flow
  * unprecise: try block holds 4 expressions. It's not clearly from code alone (we just infer from our (miss)understanding of context) where the exceptions we are trying to catch (should) happen at all
  * unprecise #2: the try block in catching the expected exceptions and also our coding bugs. There is no distinction, but it should be, coding bugs should be solved not handeled at runtime
  * no intent: intent is not clear from code, it has to be infered by the viewer (similar to #2)
  * the last except has no point in being there: What did we gain if we write code to "Expect and Unexpected error". This basicall mean we expect the block of 
  code can have bugs and we catch them print the error and raise error again? Wouldn't the interpreter print the error anyway and in consistent way? If this is 
  valid, then we should surround every few lines of our code with "try: except:"
  
## One can wish

I wished to make something that is visually and structurally not so obtrusive, it shouldn't require / create it's own code structure. I wished for something in-flow.

To have these two, the "try" shouldn't accept block of code (structure). To be in flow, an exception should a value like others, that we can handle ...

Everything below this point is just a current hypothesis, it may be wrong in parts or all-together, but let's play with it and see.

## Types of exceptions

Let's try to start from scratch ...

  * Some exceptions are a result of __programming bug__. It means the program should stop (as we don't know what will happen next and what should). Exception should be communicated to the user and logged for programmer, to fix the bug. We don't catch and handle these exceptions, we fix them if we know for them.
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
  
## The two stages of runtime exceptions in rye

Value and IO exceptions start as __failures__. Failure to do the desired operation. If failure is not handeled or returned it becomes an program error and stops the execution. So failures can happen and we can handle them, unhandled failure is an error, a programming bug.

## In what ways do we handle failures

  * we can return in to caller function
  * we can wrap it into a higher level description of error and return it
  * we can provide the alternative / default value instead of computed one
  * we can do some action (like cleanup, get alternative value ...)
  
## Handling exceptions is a translation from computer to user / domain language

All programming is or should be is a translation from computer to user / domain language, from machine code to UI basically, programming is somewhere in between.

Handling exceptions is the same. A machine failure happens and we translate it to user / domain and then display it to user. Example:
  
## Let's look at few examples from python

very simple exception, we just print to user directly:

```python
try:
    x = int(raw_input("Please enter a number: "))
except ValueError:
   print("You didn't enter a number")
else:
   print("I raise by 100 to %d" % (x + 100) )
```

```rye
   input "Please enter a number:" |to-int
     |fix-either 
       { "You didn't enter a number" }
       { + 100 |str-val "I raise by 100 to {#}." }
     |print
```

simple exception in a loop, from python docs (https://docs.python.org/3/tutorial/errors.html)

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
   input "Please enter a number:" |to-int
     |fix-either 
       { print "This was not a valid number. Try again" }
       { .print-val "You entered number {#}." false }
}
```

While in rye repeats until while return of a block is truthy.


File IO and conversion to Int ...

```python
try:
    f = open('myfile.txt')
    s = f.readline()
    i = int(s.strip())
except OSError as err:
    print("OS error: {0}".format(err))
except ValueError:
    print("Could not convert data to an integer.")
except:
    print("Unexpected error:", sys.exc_info()[0])
    raise
```

^check is a "returning function" that can also return to caller. It works like this. It accepts 2 arguments, if first is a failure
it wraps it into an error created from second argument and returns to caller (exits current evaluation unit). If first argument is not
failure it returns it. 

```rye
open %myfile.txt
  |^check "Failed to open profile file"
  |readline
  |^check "Failed to read profile file"
  |strip
  |to-int
  |^check "Failed to convert age to integer"
```

Why I feel rye version is better:

  * the flow of code strictly follows the logic and error handling flow
  * the handling of errors is locationally precise and explicit
  * check is a returning function. if it accepts the error it wraps it into a higher level error and returns that thus 
    we are translating from machine to user errors in the process, and there is no loss of information, you get the whole tree.
  * printing error seems almost always stupid. If you move this code to a function and you call the function, what does it help
    you if the function you called printed something, you mush get a return information, so you can handle it on your level (even closer to user).
    And environment determines how it will message the error to user, let's say it's a server, a phone app. In first case it would log it, in second
    it will use UI dialog for example.

## Failures across function calls 

If we put our code in function the benefit becomes evan more visible:

```rye
get-age: does {
  open %myfile.txt
    |^check "Failed to open profile file"
    |readline
    |^check "Failed to read profile file"
    |strip
    |to-int
    |^check "Failed to convert age to integer"
}
```

There are multiple scenarios you would want to do if you counln't do age. If you want to provide an alternative / default value:

```rye
get-age |fix 0 :age

get-age |fix { ask-for-age } :age
```

If you can't provide alternative, you usually want to reraise still
```rye
get-age |^check "Problem getting user's age"
```

You can then handle this up-further (closer to user), or the system displays it to the user, with nice nested info:

( Problem getting user's age ( Failed to open profile file ( myfile.txt doesn't exist ) ) ) 

## Failures aren't just strings

In the examples above I used strings to quickly create failures. But this isn't ideal, for example what if you want to use code in 
an application in another language. There are standard "error codes", I are still determining which standard to use, and there is a short-name
option that makes them translatable then.

## The pointlessness of catch and print 

As I look at the examples for exceptions in languages most of them catch and print the error. These are just examples, but I am not sure if such behaviour doesn't 
then extend into real code. All in all I think it's cumbersome model. If you don't handle (provide alternative or translate the exception) - what are you then even doing writing code?

It uses a lot of code to create user level inconsistent presenting errors. Each "app" should have one way of presenting errors determined on app level and if 
you just catch / print and fail there is no point in catching except signaling to your future self that you are aware failure can happen somewhere (you don't handle it, but you still want to distinguish it from failure you didn't expect at all (which means you must look at and figure out what to do)).

## A little bigger scenario

__Scenario: load multiple files, in their own function, translate error messages__

scenario goes like this (I have written scenario before I started writting any code to solve it)

  * load-user-name: load and read file, if error returns "anonymous"
  * load-user-stream: load concat two files return, returns string or error wrapped into "error loading user stream"
  * load-all-user-data: combine those two strings to json , if error happens at any of them return the error as json
  * result, if all works is JSON ```{ "username": "Jim", "stream": [ "update1", "update2" ] }```

TODO -- add does function

```rye
load-user-name: does { read %user-name |fix "Anonymous" }

load-user-stream: does { 
  read %user-stream-new
    |^check "Error reading new stream" 
    |collect      	  
  read %user-stream-old 
    |^check "Error reading old stream"
    |collect
}

load-add-user-data: does {
  load-user-name |collect-key 'username
  load-user-stream |fix-either 
    { .re-fail "Error reading user data" |collect } 
    { .collect-key 'stream }
  collected |to-json
}
```

The aproximate python-like code. 

```python
def load_user_name ():
  try:
    return Path("user-name").read_text()
  except:
    return "Anonymous"

def load_user_stream ():
  stream = []
  try:
    stream.append(Path("user-stream-new").read_text()) 
  except:
    raise "Error reading new stream" 
  try:
    stream.append(Path("user-stream-old").read_text()) 
  catch fileError:
    raise "Error reading old stream"
	
def load_add_user_data ():
  data = {}
  data["username"] = load_user_name()
  try:
    data[stream] = load_user_stream() 
  except:
    return json.dumps({ "Error": "Error loading stream" }) # can we get nested error info or just latest?  
  return json.dumps(data)
```

What I like about rye-version of code above

  * code flow: rye's error handling is in flow and I think doesn't disturb (visually or structurally) it more than it needs to. Try/catch is more like
    goto statements and labels
  * __intent__: rye's error handling expresses intent much better than general try/catch, fix/check/disarm/fix-else/fix-either like map/filter/reduce 
    expresses intent where for-each loop to acomplish the same doesn't.
  * functions like ^check automatically nest the errors, while I think python's usual error handling overwrites previous ones (you loose information 
    you already had). Exception handling to me (and to go-s view) is like programming about translating from computer specific to app / user specific
  * rye-s code is more symetrical, without temperary value sprinkeled all over and shorter
  * In rye all these error handling functions are library level functions, meaning you can make your own or additional for your cases
  
