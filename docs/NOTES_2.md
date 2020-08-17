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

TODO -- add read, collect, collected and collect-key

I would try to rewrite this in python like language, but franky it seems it would be quite complex code

	load_user_name: def ():
		try:
			return read("user-name")
		catch:
			return "Anonymous"

	load-user-stream: does {
		stream = []
		try:
		  append(stream, read("user-stream-new")) 
		catch fileError:
		  raise "Error reading new stream" 
		try:
		  append(stream, read("user-stream-old")) 
		catch fileError:
		  raise "Error reading old stream" 
	
	def load_add_user-data does {
		data = {}
		data["username"] = load-user-name()
		try:
		  data[stream] = load-user-stream() 
		catch:
 		  return to_json({ "Error": "Error loading stream" }) # can we get nested error info?  
		return to_json(data)
	


