> Notes since the project went on github

# 08.08.2020

## few vocabulary decisions, loader

While writing intro_2 I decided I need to implement some things that move us more away from rebol (not the core at all, but how it worder the things on top, 
the "frontend") to be more concise and in line with rye 

 * For typical dyadic operations, where known operators exist they shoulb be prefered over words (a much used prefix _join_ in rebol becomes +, 
   greater? and lesser? become > <). In rebol prefix was preferred, but we use much more infix words anyway 
 * join becomes like python's and javascripts _join_, 
 * joining a block without separator should then be _concat_ intead of rebols _rejoin_ as you don't really _REjoin_ anything when you do that. 
   concatenate is a little to long and complex word for this very used function, so concat should be one of limited few words that we adopt as our jargon
   
the idea is that the opeators are nothing special, they are functions like all others, and are recognised as opwords, and they don't have a regular non-opwordy 
counterpart ... although if a benefit would show they could have it like > is opword (infix)  _> is word (prefix)
I need to update the loader for this.

     if greater 100 10 { print "100 is greater" }
     // is the same as 
     if 100 .greater 10 { print "100 is greater" }
     // will be the same as
     if 100 > 10 { print "100 is greater" }
     // could be the same, if any benefit shows up as
     if _> 100 10 { print "100 is greater" , 
  
Logic should probably be that all one character operators are by sefault recognised as opwords where naming isn't one dot less, but one line more?

  opword > ==becomes==> name _> so setword in this example would not be:
 
     >: fn { a b } { greater a b } // but
     _>: fn { a b } { greater a b }
  
## Rebol and dictionaries

Rebol was made before we began using dictionaries for all sorts of things in dynamic languages (or objects in JS). I was writing some rebol code this week
when I needed to use a simple dictinary like structure, to sum by keys, and either I forgot about how to do this in rebol or it's very cumbersome. It's not like
rebol couldn't do this but because of it's syntax the shorthands we are used to 
  
     a[key] = a[key] + v
     // well I tried now some more and you can do it pretty much the same, but I still don't that much like that double colon ... it's basically an accessor, with a 
     // get word while all around being like a set-word in one ... :P, but it is comparable to py/js code, which I complained about
     b: [ "a" 100 ]
     c: "a"
     b/(:c): b/:c + 11
     b/:c: b/:c + 11 

I don't know if there is something better and as consise, but we have to explore this.


Another thing I want to change about dictionaries, hasthatables, ... in Rebol and now in rye you can only type-in blocks, which is nice unary thing.
And they can then be turned to anything by functions. In rebol

     block: { "a" 1 "b" 2 } // this is a block, if you want hashtable performance you do
     block: hash { "a" 1 "b" 2 } // but this loads a block and then a function turns it to hashtable. If we have a big hashtable, we want it to be loaded
     // directly as hashtable

Third. Because block and dictionary are not distinguished, you can use all block/dict functions on both, which is "maybe??" nice, but dict has a very specific 
simple API (get/set) and it's probably more a source of errors and confusions. And block has a very wide, flexible api .. just few Examples ... 

     a: [ b 1 c 2 ]
     select a 'c // returns 2 is like a/c , but select works the same on key and value arguments
     select a 1 // returns c which is really weird in dictionary

     finda a 'b // vs. 
     find a 2

With dictionary you always know if you are calling something based on key or value ... I think it's never a benefit to have these two undistinguished. What if
keys and values are for example names of people 

     married-people: hash [ "jane" "jim" "paul" "natasha" ] // who is married to who if we fo _find_ for jane and for paul

I think we need dictionary not be just different rye value internally, but also in code, so it needs it's own syntax which loader can recognise. We have {}
for blocks of code and blocks (lists/ arrays) in general. Current idea is to have [] as a specific block. It could be just dictionary, but maybe some internal
load time rules could be used to enable other types too ... like sets .. maybe

     some-dict: [ a: 10 b: 20 ]
     some-set [ a 10 b 20 ] 
  
Questions: 
* But what is fist key now: 
 * a string "a"
 * a word _a_
 * a set-word _a:_
* Are sets really that usefull in practice?
* Can dictionaries (or sets) be used as specific code elements in which case we want that they can include all rye values
  * we can and do use dictionary like blocks as code, in specific dialects, but they internally can't be and aren't  woudictionaries
  * are dictionaries evaluated implicitly or explicitly, just values or keys also? with compose reduce ... 
  * ...
  
The compose route ...
  
     val: 12 key: "ab"
     some-dict: compose [ ab: "val" (key): (val) ]    
    
If dict can hold only literal values, they we could eval all values, and only the keys that are explicitly marked like
    
     some-dict: [ ab: "val" ?key: val ]
  
I am begining to understand why rebol would rather have just blocks ... but I still think we need them on a practical side ... need to think about it more.

Overusing dicts for programming ... "dict based programming" is by default slow and also more error prone. I would much better prefer something like records 
in ocaml or structs, but I am not sure if you can even make (or reap any benefit) from making a "static" structure in a dynamic (runtime) language?

Whole subject on dicst is something to ponder on and try things ... One goal of rye is to make more dynamic / flexible language, but also more exact when you 
want it to be exact. Exact dicts with exact api, would be one step towards it. On the implementation side ... would dictinaries be just objects (tuples) or 
are there any differences?

## loops with injected values vs ordinary

in rebol we have for-each

      as: [ 1 2 3 ]
      for-each a as [ print a ]
      
Rye won't have nonevaluating word exception determined by the caller func so this would be
 
      for-each 'a as { print a }
      
But rye also has injected blocks, which play well with opwords, but do we want to force everyone to use them, or always use them:
      
      for as { .print }

sometimes we dont use pipe flows, we need a repeated access to value in such time we do

      for as { :a print a }
      
 same for map, filter, reduce and other similar
 
      map as { .add 100 }
      
but this is probably less elegant or at least less __usual__ for programmers from other languages so we could have equivalent *-each functions
work 

      for-each 'a as { print a }
      map-each 'a as { add a 100 }
      
# 10.08.2020

## Operand op-words

Commit for one-letter opwords.

But they will not be just one letter op-words like "<" ">" "=" "+" "*" ... change parse rules so we also include likes of "!=" "*=" "==>" and other combinations.
These all are automatically loaded as op-words.

## Collect besides return
 
in general evaluation of blocks returns the result of the last expression in a block. You can change that in special cases with return function.
I was thinking about something like this for a while ... since block is a very universal collection type in Rye ... what if we have a collect function, that
can if called create a (code) block level blok to which it appends. And at the end of block if anything was collected then this is returned unless of course explicit return is called.

      probe do { print collect 1 print collect 2 }
      1
      2
      { 1 2 }
      
 this could get more usefull in loops and various state machines, parse dialects ?
 
      a: 0 probe loop 5 { a: collect inc a }
      { 1 2 3 4 5 }
 
 
# 14.08.2020

## Some thoughts on kinds

I was looking at ruby example on front page, trying to translate it to kinds.

Since methods are not part of the kind (just data) and we have generic methods ... first example looked really clumsy compared to Ruby

```ruby
# The Greeter class
class Greeter
  def initialize(name)
    @name = name.capitalize
  end

  def salute
    puts "Hello #{@name}!"
  end
end

# Create a new object
g = Greeter.new("world")

# Output "Hello World!"
g.salute
```
rye so far

```rebol
def-kind 'Greeter {
   name: required string calc { .capitalize }
}

def-method 'Greeter salute {
   >name .printo "Hello {#}!"
}
   
<Greeter> { name: "world" } |salute
```

The repetition of Greeter at method is not the best. I am thinking if the kind definition dialect would besides being a validation didalect also accept
function definitions and set the generic methods (they are still not part of the kind, but can be defined nested in definition and outside)

I also think ... what if we just accept the "class" word instead of using some third, "kind". To add to familiarity not take away. I was thinking that "class"
is too overloaded with meaning and expectations, but rye does a lot of familiar things a little differently. Let's try it for a while 

```rebol
class 'Greeter {

  name: required string calc { .capitalize }
  
	salute: does { 
    >name .printo "Hello {#}!" 
  }
}

methods Greeter {
  chow: does {
    print "woof"
  }
} 

<Greeter> { name: "Jim" } |do { .salut , .chow }

{ name: "Jim" } >Greeter> .salut
```

if the kids definition dialect would not include method definition (just validation) then we could do it like this,
but it's still visually heavier

```rebol
class 'Greeter {

  name: required string calc { .capitalize }

} .methods {

  salut: does { 
    >name .printo "Hello {#}!" 
  }

}

methods Greeter {
  chow: does {
    print "woof"
  }	
}
```

# 17.08.2020

## Exception handling in light of file IO

I've added some file IO functions. So I can now try the failure/error ideas in more practical scenarios. I have a feeling
they will work for some cases, but sometimes, traditional try/catch structure will still be preferable (for many consequent IO operations, 
you don't want to handle individually)

### scenario 1: reading file and printing it's contents

So we start with reading a file and printing it's contents

```rebol
open file://test3.txt |read-all |print
<Error: Word not found: <Word: 92, read-all> >
```

If the file exists it all works, if not it currently returns <read-all word not found> as read-all is only determined on kind rye-file, and
we returned an error kind. First, maybe error text should be more explicit. Maybe we should check if word exists at generic functions list and 
report what kinds could be ok. Maybe 
	
Basically this is first of all a bug in raising errors conencted to pipe words. If we do
	
```rebol
open file://test3.txt :a read-all a |print
failure
critical-error
<Error: open test3.txt: no such file or directory >
```
Then error makes sense. We should first fix this bug in the interpreter, as we want to mostly design a in-flow (pipewords) exception handling
with escape words.

Found the bug ... I need to go through all this code again, and make the loops with words opwords / pipewords clearer. Interpreter checked if the
next word will handle the failure, and this worked ok if it found the next word. Otherwise it overwrote the first failure for failure for not found next word.

Ok so now we have: 

	{ Rye } a: open file://test3.txt |disarm |print
	<Error: open test3.txt: no such file or directory >
	{ Rye } a: open file://test1.txt |disarm |print
	<Native of kind ryepr-file>
	{ Rye } a: open file://test3.txt |^check "Problem opening the profile file." |print
	Failure
	<Error: Problem opening the profile file. <Error: open test3.txt: no such file or directory >>
	{ Rye } open-profile: fn { } { a: open file://test3.txt |^check "Problem opening the profile file." |read-all }
	<Function: 0>
	{ Rye } open-profile 
	Failure
	Critical error:
	<Error: Problem opening the profile file. <Error: open test3.txt: no such file or directory >>
	{ Rye } open-profile |print
	Critical error:
	<Error: Problem opening the profile file. <Error: open test3.txt: no such file or directory >>
	{ Rye } open-profile |disarm |print
	<Error: Problem opening the profile file. <Error: open test3.txt: no such file or directory >>
	{ Rye } open-profile: fn { } { a: open file://test1.txt |^check "Problem opening the profile file." |read-all }
	{ Rye } open-profile |disarm |print
	profile-data ...

This make sense, except ... in repl, later figure out what does the returning function do, does it just return the failure or should it 
raise critical error.

OK, so back to initial scenario ... can we use simple fix ? In this case we can't really ... 

	{ Rye } a: open file://test3.txt |fix "no data yet" |print  }
	no data yet
	<String: no data yet>
	{ Rye } a: open file://test1.txt |fix "no data yet" |print  }
	<Native of kind rye-file>
	<Native of kind rye-file>
	{ Rye } a: open file://test1.txt |fix "no data yet" |read-all |print  }
	soso
	<String: soso>
	{ Rye } a: open file://test3.txt |fix { "no data yet" } |read-all |print
	<Error: Word not found: <Word: 95, read-all> >
	Critical error:
	<Error: Word not found: <Word: 95, read-all> >

Read-all should only be applied if error didn't happen. We need something like

	{ Rye } a: open file://test3.txt |fix { "no data yet" } { |read-all } |print
	
Question: should fix accept literal value or a block to execute. A literal value is shorter but has limited use, expressions are problematic, since we are allready  in failed state and interpreter as it is now doesn't want to accept new expressions ... just the first one ... if we accept a block we can change the state, and then evaluate it. What would the name be for either case ... fix { } and correct { } { } doesn't really make real sense. Maybe fix and fix-both, (fix2, either-err, errther...???

	{ Rye } a: open file://test3.txt |fix-both { "no data yet" } { |read-all } |print
	
Made the fix native funtion accept block with, and added fix-both so this works now:

	{ Rye } open file://test3.txt |fix-both { join "jo" "jo" } { |read-all } |print
	jojo
	{ Rye } open file://test1.txt |fix-both { join "jo" "jo" } { |read-all } |print
	profile data ...

... how / when do we close the file ... can we use defer as in go ... defer being a injected block function. But we can't put .defer { .close } before
fix-both ... and we can't put it after. :P 

So one way now seems a try function.

	{ Rye } open file://test1.txt :file |fix-both { join "jo" "jo" } { |read-all } |print try { close file } 
	profile data ...

If we wanted to do it all in stream we have one problem ... unless we invent some "faliure passing function", skip doesn't yet solve the failure, but it does
pass it forward without triggering error ... in a way it could work since skip evaluates a subblock and in subblock a first word does dissolve failure
or we just add disarm for now. we could change fix to work on type of argument not on the flag

	open file://test1.txt |disarm |skip { 
	  .fix-both { join "jo" "jo" } { .read-all } |print
	} |try-w/ { .close }

one way withouth this would be

	open file://test1.txt |fix-both { join "jo" "jo" |print } { .read-all |print , .close }

since skip returns an error ... we could use the fix-* also ... this works now. We could add something like fix-else

	{ Rye } open file://test3.txt |disarm |skip { |fix-both { join "jo" "jo" } { |read-all } |print } |fix-both { } { .close }
	jojo
	{ Rye } open file://test1.txt |disarm |skip { |fix-both { join "jo" "jo" } { |read-all } |print } |fix-both { } { .close }
	soso

We could do the same with commas, or with ... depending on what we wanted returned

	open file://test1.txt |disarm |with { 
		.fix-both { join "jo" "jo" } { .read-all } |print, 
		.fix-else { .close }
	}

Seems nicer ... TODO -- add with native func. It's the same as skip, but it returns the result of last expr. not the first arg.

Ok, native functions have "AcceptsFailures" flag and we added this flag to with (and skip) so this becomes (it makes sense for this flag to be on
for "combinator" functions), with changes to else block we don't need with basically.

	open file://test1.txt |with { 
		.fix-both { join "jo" "jo" } { .read-all } |print, 
		.fix-else { .close }
	}

	open file://test1.txtt fo
	  |fix-both 
	    { "jo" + "jo" } 
	    { .read-all :text , .close , text } 
	    |print

Returns function would make it even shorter. Will probably try adding later. 

How would this look in something like python?

	text = ""
	try:
	  f = open("test1.txt")
	catch:
	  text = "jo" + "jo"
	else:
	  text = f.readAll()
	  f.close()
	print text

Offshoot: Is there some combinator that would handle this pattern without returns? What would it be like:

	100 { return-first { .add 10 } { .print } } 

	open %test1.txt
	  |fix-both 
	    { "jo" + "jo" } 
	    { keep { .read-all } { .close } } 
	    |print

Keep would do inject 2 blocks with it's first argument, but keep / return the result of first block. Not very obvious and more
verbose than returns solution it seems ... we can keep it in mind

TODO .. lets add %text1.txt as shorthand for file:// to loader. and .returns function

	open %test1.txt
	  |fix-both 
	    { "jo" + "jo" } 
	    { .read-all |returns , .close } 
	    |print

Maybe keep is not so bad, it's more symetric and with all the injected block behaviour it could be quite poverfull in situations. Fix-both maybe 
visually belongs closer to open as it's relating to it. So the final version would be. skip could maybe be better named "pass" as it's passes it's value
over block, or passes over block.

	open %test1.txt |fix-both 
	  { "jo" + "jo" } 
	  { keep { .read-all } { .close } }: 
	  |print
	  
### Scenario 2: load multiple file, in their own function, translate error messages

scenario goes like this (I have written scenario before I started writting any code to solve it)

  * load-user-name: load and read file, if error returns "anonymous"
  * load-user-stream: load concat two files return, returns string or error wrapped into "error loading user stream"
  * load-all-user-data: combine those two strings to json , if error happens at any of them return the error as json

TODO -- add does function

```ocaml
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
    { .^check "Error reading user data" } 
    { .collect-key 'stream }
  collected |to-json
}
```

TODO IMPLEMENT -- add read, collect, collected and collect-key

I would try to rewrite this in python like language, but franky it seems it would be quite complex code

```python
def load_user_name ():
  try:
    return read("user-name")
  catch:
    return "Anonymous"

def load_user_stream ():
  stream = []
  try:h
    append(stream, read("user-stream-new")) 
  catch fileError:
    raise "Error reading new stream" 
  try:
    append(stream, read("user-stream-old")) 
  catch fileError:
    raise "Error reading old stream" 
	
def load_add_user_data ():
	data = {}
	data["username"] = load_user_name()
	try:
	  data[stream] = load_user_stream() 
	catch:
	  return to_json({ "Error": "Error loading stream" }) # can we get nested error info?  
	return to_json(data)
```

### TODO -- improve this code, make it realistic also with concrete modules / functions

What I like about rye-version of code above

  * code flow: rye's error handling is in flow and I think doesn't disturb (visually or structurally) it more than it needs to. Try/catch is more like
    goto statements and labels
  * __intent__: rye's error handling expresses intent much better than general try/catch, fix/check/disarm/fix-else/fix-either like map/filter/reduce 
    expresses intent where for-each loop to acomplish the same doesn't.
  * functions like ^check automatically nest the errors, while I think python's usual error handling overwrites previous ones (you loose information 
    you already had). Exception handling to me (and to go-s view) is like programming about translating from computer specific to app / user specific
  * rye-s code is more symetrical, without temperary value sprinkeled all over and shorter
  * In rye all these error handling functions are library level functions, meaning you can make your own or additional for your cases
  
  
### The fail of: Catch, (reword) and print
  
Many languages oftex exhibit error handling in manner catch and print. You catch the error and then you print the error. It creates code and structure and
acomplishes very little, and it's ununiform. My thought are, that things can be much more thought out. There is a number of common scenarios of what you want to
do, and they should be thought out in advance and systemized on language and app level.

#### Failure you expect and will handle (continue the program)

Some failures can be expected to happen and you want to handle, continue the application, redirect the logic, provide alternative / default value ...

#### Faliure you can't didn't expect but happened

Everything that happens but wasn't expected, __should stop the execution of program__ as you don't know what can be the consequences. The failure should 
be printed out / displayed and / or logged. It's a bug that needs to be solved.

#### Failures you expected but can't handle

If you don't intend to continue running the program then there is probably little reason to handle them specifically, maybe only to mark that it's a know
potential failure that you dont / can't handle. So you don't treat it as a but that must be solved. 
re is 
The best solution here seems wrap original failure with this additional info and let it stop the program. These as previous types of program should 
print something to the user, not just silently crash. But it needs to be determined systematically by whole app the same as previous category.

Solution to both these is that you can on app level determine a global handler.

	
```rye
	set-app 'on-error { e } { probe "Unexpected e" + e } // log e , dialog.alert("Error happened:" +\ e) ....
```

# 25.08.2020

I posted the scenario 2 from last notes on fb and we @sbelak was interested to see how we dispatch on types of errors if we can. I see it as a good next scenario
to think about.

## Exceptions (evolving) scenario 3

Let's try few scenarios, from simpler and complicating / more or less randomly changing them and see if they are solvable in a nice way. The focus is still
on flow of code using the current idea about failure management. The exact strucutre of error object is not an issue yet.



load a page and print it's content or print the custom error description


```rye
get https://www.google.com 
  |fix-switch {
    404 { "stran ne obstaja" }
    403 { "nimaš dostopa do strani" }
  } |print
```

ok, now we want to do something with the OK result, and still print the error messages


let's say we want to print the length

```rye
get https://www.google.com 
  |fix-either { 
    .fix-switch {
      404 { "stran ne obstaja" }
      403 { "nimaš dostopa do strani" }
    }
  } { 
  .length? 
  } |print 
```
we want to get the length and check if it's above 1000 and return bool, in case of error we still want to print the custom message and return the error

```rye
get https://www.google.com 
  |fix-either {
    .pass {
      .fix-switch {
        404 { "stran ne obstaja" }
        403 { "nimaš dostopa do strani" }
      } |print 
      }
    } {
    .length? > 1000
    }
  }
```

ok that printing makes no sense ... let's return 200 if all is ok and the status code if it's not

```rye
get https://www.google.com 
  |fix-either {
    >>code
  } {
    200
  }
```

// note: >>accessors are not yet made

ok ... let's try to do some actions based on error types and if all ok save page to local file

```rye
notify: jim@example.com

get https://www.example.com 
  |fix-either { 
    .fix-switch {
      404 { .re-fail "Can't access webpage" }
      402 { .to-text >> 'body <Email> { to: notify subject: "Pay for page please" } |send }
      403 { .to-text >> 'body <Email> { to: notify subject: "Access was forbidden" } |send }
      }
     } { 
      .write %the_page.html
    }
  }
```

// TODO to make these examples work we need a, but accessors and constructors are in plans after we finalize kinds ... I will see
// Maybe I will try to make something before.

  * get generic function that works on http-schema
  * fix-switch that can switch on status codes (later words too)
  * tuples , constructor and set operator >> 
  * write function for files
  * send function that sends email based on <Email> tuple


# 29.08.2020

## Exceptions - first shape

We have 3 handling phrases, all 3 return it's first arg if it's not an error: 

  * check - checks if first arg is error, if it is it rewraps it in (new error) second arg and returns that
  * fix - checks if fist arg is error, if it is it executes code (second arg) and returns the result of code
  * tidy - check if first arg is error, if it is it executers code (second arg) and still returns the error

all these phrases have multiple modes besides the basic one

  * -switch switches on type or code or error 
  * -either accepts is-error and isn't-error block of code

## Emails and ~/.rye-profile and interesting options

Added support for send function. It's meant to very quickly send a email notification to someone. I added rye-profile file.

If it's there in the user's home directory rye does it. In this case it sets a context user-defaults where also smtp info is set. This will get tuned, named better. We would need to have an option to ignore per "project" (directory) the use of this file and at the same time enable per-project defaults.

This could be achieved if rye looked at current directory for a file .rye-project and load it and that one would also be able to disable the loading of user-profile. Would there also be a web-scope possible, would it make sense?

This would also make a directory some sort of app if it loaded all the code in the rye-project file. From web or from local files. You could just create a folder and add a rye-project file directing to online file and all would be loaded (with caching) and started.

Especially in combination with Spruce based console UI. ;)

## Some ideas

### Weird idea of "drop-in - drop-out"

I got an idea last time. If I could call a word drop-in that would instead of moving to next blocks call a console and let you do repl inside any point of code.
good idea would be if you could then also drop-out and continue the evaluation normally. Even better if you could change the current block and replay it multiple times and then drop out. I am not 100% sure this is doable (I have a feeling there must be some catch 22 otherwise others would have done it), but from very cursory look I don't see why not.

### The try { retry } failed idea

I had some idea that try woul record it's state, then enter code and if retry was called inside it would reset the state to saved and rerun the code in try. It failed soon after (idea) because, this would be a very good producer of infinite loops. Maybe if there was retry-once ... or try 3 { .... }. But there are also other issues like the timing of retry ... etc ...

## Example works

So I added email value type, send generic function, get, post generic and tidy-switch on statuses

    email: refaktorlabs@gmail.com

    get https://example.com/data.json 
      |^tidy-switch { 
          404 { send email "data is missing from site" } 
          _   { send email "site is not working" } 
        } 
      :data 
    post https://httpbin.org/post data 'json

## Next!!! Failure management + Validation + Kinds !

This is where it potentially all comes together, or falls appart at least somewhat. The combination of kinds (generic words on them), the validation dialect, failure handling from that validation. If I can make this all 3 work in sensible it would be very good, otherwise I will need to rethink some parts. oon

These will also be a lot of deciding to do. When are blocks of kind validated, allways? is there a flag. Can it be lazy ... when the reciever requires it? Is there special syntax for creating a validated object

What would be the inital example we start working with. While we are at send ... send will also accept Email kind object. This would be a start ...

If we want to send a simple email with the object:

    <Email> { to: refaktorlabs@gmail.com
              subject: "Hello from Rye"
              body: "yup ..."
    } |send
    
Let's say we get the json data from some service, let's say the constructor validates always:

### Scenario #1: fail on validation

    get-data >Email> |check "email data didn't validate"
    
### Scenario #2: do something on validation

    get-data >Email> |tidy { alert "email data didn't validate" }

### Scenario #3: if email wasn't correct ask user for new email

    get-data >Email> |fix-switch { 'to { .prnf "Entered email was {#}, enter correct one: " d << 'to input >Email>  }
    
There is some unelegant repetiton there and it only works once, then if email is still not ok it fails with error. Maybe that try-retry wasn't so bad idea ..
or we need another construct that can modify values and redo the operation implicitly? ... 

    try { >Email> } { fix-switch { 'to { .prnf "Entered email was {#}, enter correct one: " d << 'to input } } }

closer .. but what about combined cases, etc ... think more .. should be something implicit in constructor maybe?

    try-make 'Email { .. will keep retrying with result of this block .. }
    
Think of more realistic interaction scenarios ...

## Let's look at it from different view

Look at cases when we would like to validate and how to handle.

### we load data from some server

if we load data, we can't reply, we can log / store the error (by evolving it and letting it fail). The error will log, we will see the hiher description
and the basic error, which describes the exact reason validation failed.

    laod-data-from-server >Email> |check "data didn't validate" |do-with-data

### we are the server, we get data via request

we respond to the client with the validation error as JSON. Our server function works on business logic and it's reply

    get-post-data >Email> |^check none |do-with-data |get-response  

    get-post-data >Email> |^fix { .to-json } |do-with-data |get-response |to-json  

### the direct user or GUI app 

we show the user problem directly in gui

    get-post-data >Email> |^fix { .display-validation-errors } |save-file  

# 25.10.2020

## What I want to think about

* **Exceptions as result of validation dialect**. In case of validation exception has child nodes that can be key or index related, or nested deeper . Are those nodes also exceptions or something else?

* **The billion dollar mistake**, runtime null/exceptions code errors. Option and Either types make more sense in compiled languages, where you can statically, at compile time prevent null-exceptions. Dynamic and non-compiled languages don't have this benefit, but they could still improve *certainty*  from these constructs.
A interpreter should issue (log) warnings or fail because of type error when it mees a code where null-exception is possible. The difference is that it fails/warns every time it sees this could be possible with given code, so you fix these cases. Not that it fails in specific case when null really happens. If we
could make handling of option types elegant this would be a good feature to have. Same thing for error values.

## Exceptions and validation

Some terminology (current proposal). Exception is an object / rye value. Failure and Errors are a state of interpreter. If failure is not handeled it becomes and error. It would terminologically speaking make sense for functions to return exceptions withot failure too?
 
### Exception structure

Exceptions have: "custom message" (string), code (int), type (word), parent(parent exception), parents (List of exceptions).

(we have a separate parent / parents) because most times there will be single parent and we don't want to create lists with one parent for no reason)

Exception construction dialect accepts those 3 types or block of those 3 types automatically to construct exceptions. 

fail 404  ; fail "user wasn't logined" ; fail { 404 missing }

In case of validation exceptions the children also need "key" and/or "index" { validation-fail { { name: required } }

Key could be a setword in constructor, for index integer clashes with code ...

(idea) codes except few typical ones like 404, 503, 200 are a queswork anyway ... if we have types, they should map to codes, but we should use words anyway,
not numbers, even in case of 404 "not found" is better and more descriptive.

Ok so we ditch code and integer in constructor is index. So far the exception doesn't hold any information regarding file/line of code where it happened.

Revised structure: 

	{ type "custom message" <Integer(index): 0> <Exception(parent)> <list: ().... parents ....> }

### So how do we handle validation exceptions again?

** case 1: on exception return genaral wrapped exception **

	person: { name: required age: required integer }
	dict { "age" 33 } |validate person => "name"  // code error: validate didn't return a dict but an failure
	
	// if failure just wrap and return, otherwise return name
	dict { "age" 33 } |validate person |^check "person data is invalid" => "name" // code ok

	// if failure print something, otherwise print name
	dict { "age" 33 } |validate person |fix-either { print "person data is invalid" } { => "name" |print } // code ok

	// if failure print the keys and messages, otherwise print name
	dict { "age" 33 } |validate person |fix-either { -> parents |for { -> 'key |prn , -> 'msg |print } } { => "name" |print } // code ok

	// since it's general pattern to do something for parents in validation scenario we could have the specific function for it
	dict { "age" 33 } |validate person |fix-parents { -> 'key |prn , -> 'msg |print } { => "name" |print } // code ok
	
	// we would need eihter there !? this becomes combinatorial explosion ... so what if instead of -either we by default use ^
	dict { "age" 33 } |validate person |^fix { -> parents |for { -> 'key |prn , -> 'msg |print } } => "name" |print // code ok
	dict { "age" 33 } |validate person |^fix-children { -> 'key |prn , -> 'msg |print } => "name" |print // code ok
	
	// but what if we can't return. if there is other code to execute after failure anyway, maybe we split the problem in block
	dict { "age" 33 } |validate person |fix-either { .for-children { -> 'key |prn , -> 'msg |print } } { => "name" |print } print "bye" // code ok

	// a more functional approach
	dict { "age" 33 } |validate person |fix-either { -> 'children |map { .embed "{{key}}: {{msg}}" } |print-lines } { => "name" |print } print "bye"

** what if we want to do something in case of specific field **

	// some of these maybe don't make sense as it would be solved on other level, but let's say we need to somehow
	// if failure is with name we show a poput to user to enter his name, syncroniusly
	dict { "age" 33 } :p0 |validate person |fix-either { .match-children { name: 'missing { prompt-user "What is your name again?" .update p0 "name" } } } { => "name" |print } print "bye" // .... let's retry that
	
	
	dict { "age" 33 } :p0 |validate person |fix-children { name: 'missing { prompt-user "What is your name again?" |update p0 "name" } } => "name" |print

	// this is fail, we didn't validate the second name the user entered the correct code would be (where we ask and if it fails again we fail
	dict { "age" 33 } :p0 |validate person |fix-children { name: 'missing { prompt-user "What is your name again?" |update p0 "name" } } 
	  |validate person |^check "We give up!" => "name" |print

	// if we are here ... can we make this circular? so if it doesn't validate it keeps asking ...
	dict { "age" 33 } |assure-valid person 'p0 { fix-children { name: 'missing { prompt-user "What is your name again?" |combine p0 "name" } } } 
	   => "name" |print


So what would we need if we this would be the way ...
  * fix-children with special match dialect, that uses the return of block to feed to next match
  * assure-valid which is sort of loop until validation doesn't return error
  * update / combine ... takes dict, overwrites the value and returns (new) dict, since it's single key maybe we should just call it "set"
  
**The UI code**

In practice, in real apps we usually don't fix most of the validation errors but return them to the UI / frontend layer as it is and it displays them 
to the user. This above is all more to test if this is flexible. In practive there is a separate function that displays the error, we don't adhoc display them
inside fix code.

**So what are our conclusions for now**

Handling validation errors is no different in general than any other erros. Basically we use the check, fix and tidy, and we could have more special functions
for specific cases, but we would need to find those cases in the wild. There doesn't seem to be any that typical use to make them beforehand. Maybe fix-children, but even there I am not sure if I will really need it exactly as that.

# 26.10.2020

## Next per example

So if I want to follow the example from the first blogpost, next thing to do is:

https://ryelang.blogspot.com/2020/10/blog-about-rye-language-development.html

 * make load generic function for kind scheme: it should look in local cache and if not there save from url to cache and load
 * url should accept the {{variables}} embeddings and embed them, probably at the same time add it to strings. Didn't yet decide if it should work by default
   or if there is a function required. On first sight it seems that function is better, if so then loader should also accept such string/uri and
   only the embed function should find and replace those. Otherwise probably the loader should recognise them and parse them out. This would make all string 
   or urls more expensive to create and heavier / more complex?
 * on-error callback
 * validation dialect should return validation failure when validation fails, with children detailing the reason
 
## Other paths forward

To further explore the validation / failure mechanics it would be good to make something like a web-server. I already have golang's Echo builtins made, but
I was thingking standard library webserver is quite powerfull in go anyway, so that should be the defatul basis for providin a webserver functionality in 
rye. Echo is an optional addon.

Another interesting path towards Rye-s toolset for distributed computing options would be nng which is nanomsg which "was" zeromq. nng supports tls protocols
for example so it's meant to be for external / exposed comm also, not just internal. Making direct actors/mailboxes for distributed computing seems simple at 
first sight, but you have to build all the specific plumbing on top of it. nng also provides the typical distributed patters, the actors and messages can probably 
still be there, but messages delivery has more options by default than just a2a. So it would be interesting to start playing with nng. Also test the ideas
of "mobile code" I was thinking about. For mobile code one thing we would need would be distinction of pure and unpure functions. ..>

## Pure functions

The idea is that we would specify pure builtins and user can create pure function by just using pure builtins or functions. These functions can then be more mobile (in relation to previous topic), since they are safer to use as a filter / seek / map operation on some distant node. Pure functions of course have other
guaranties that improve reliability of local use.

### Implementation

"Pure function" would need to be a flag in a builtin and function. But we want to avoid need for checking this flag at evaluation all the time. One solution
would be that we create a context of pure functions (they could be available in both ordinary and pure context) and all pure user functions are bound to that
context, not the ordinary one. So unpure function/builtin would be undefined in a pure function code. I need to try this solution, we already have very flexible
context manipulation (look at isolates), so this could already be possible inside a language almost.

# 15.11.2020

## Http server + websockets

In the meantime I made a stdlib-s golang's http server in rye. I also added a websocket handling. Since the stdlib is said to not be that good I used 3-rd party. There are at least 3 
contenders here. Maybe I will switch to another one as I dive deeper into this. Currently the stupid problem is that the error in the handler isn't reported. Even worse, the websocket handler
stops reponding. And even worse still, after we close the websocket client connection handlers stop responding also (I think they didn't initially, so I maybe introduced some error on stop). 
It's hard to tell since we don't get any error reporting or printout from the handler.

	new-server ":8080"
	 |handle "/" "Hello from Rye!"  
	 |handle "/fn" fn { w r } { write w "Hello from Rye function!" }  
	 |handle-ws "/echo" fn { s ctx } { forever { read s ctx :m , write s ctx m } }  
	 |handle-ws "/ping" fn { s ctx } { forever { read s ctx :m = "Ping" |if { write s ctx "Pong" } } }   
	 |serve

## on-error-do

So this should be the next thing to tackle. After a little thought I saw why the error isn't printed out. As it is now the repl prints out the error. And in case of a hadler there is 
no repl interaction involved. The interpreter itself doesn't do this, because I was thinking that "printing" an error is just one case what you want to happen on error. The "what should happen"
is app or environment specific. Maybe you want to log it, maybe display it in UI, send it to client, etc ... 

The idea was always that there should be an option of defining an sort of "od-error" function in your or projects profile, or code itself and it would determine this. Thinking now about it, 
it should be context specific, so you can redefine it in specific contexts. If you don't and are in a normal rooted context then the parent or root context definition is used. So defining
a normal word inside context gives us this desired behaviour it seems, So we don't need a special pointer inside context or some other environment related object.

## the interpreter should then

so on error interpreter should use current context to search for the on-error-do word. If it's not there then for now it does nothing, just dies, if it is there it disarms the error and 
calls it with error as a first argument. Keep in mind that the contexts can be nested and also errors can be nested. **Let's do this.**

[ ] on error in interpreter call a on-error-do function
[ ] some code met at runtime is a coding error ( like "do [ print ]" with no additionl arg). Is this the same type of runtime "error", or something else, a "code mistake" that needs to be fixed either way? Handle this at least temporarily.
 
## Handler function + program state

So far the handler just called a regular function, but I am not sure this is ok really. The ordinary function holds a reference to a programstate that changes. This is then shared between
the handler functions, but at least some data, like current block and the position of ealuation inside it shouldn't be. Also these functions should not have a write access to words in parent
contexts (maybe no functions should anyway ... I haven't fully determined if this is really needed :/ ... you need read access to have access to functions and "constants" that you defined in parent contexts)

In any case ... we should dissect program state and think what parts are shared in concurrent use, what are ok to be shared and what do we need to create new version of program state 
that can run in it's own "thread". Could we serialize this progrm state or swith between them freely?
