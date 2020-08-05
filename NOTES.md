# Notes at Rejy0 Go

## 23.01.2019

We started writing eval functions. Look at javascript implementation and reimplement it here unless you see a better model.
We should have EvalContext or something like it that holds the series, position, environment and return value (or maybe value stack???)
This is then passed into eval functions and returned from them. Like EvalBlock, EvalExpression, EvalFunction, EvalInteger, EvalBuiltin ...
Eval is independent of Objects / Nodes, since nodes are data and we have multiple dialects evaluating the data. Do is the default dialect

Do is also a function that evaluates a block otherwise. You could specifiy the eval dialect in Rejy header .. you could also have stack dialect for
example, or other experimental ones. That you could also load at runtime.

Edea from Ren-c ... Error is a path value

Testing general ideas we have about lang ... to see how it would come out

=== file.rejy
Rejy { "Documentation" Do }

import {
	http
	xml-dsls
}

get-links-to: fn [ url to ] { 
	read url |html->xml 
	|xml-reduce 'o {} [ on <a> 'x { append o x } ] 
	|fiter 'x [ found? find x to ] 
}

get-links-to http://www.cebelca.biz "ajpes" |probe

===
I think opwords are great :) . Posted and example of fb and tw.

Next we look at evalExpr in JS version and recreate it more or less. Also make evalBlock. I haven't yet fully thought through, if having reader/loader
nodes be the same objects as runtime nodes (stored to env, returned and consumed via functions) are is nice benefit or something wrong. It's probably 
a benefit because there is no conversion, and functions accept nodes literally from reader or from expression (i.e. functions)

I am also not sure if we could in production state if functions for example are locked just replace words refering to functions directly with references to
functions. 

    fn [ x ] { add 1 x }
  	on evaluation context is created with x binding to argument value
	add word references to builtin, which is dereferenced and then called
	if (add word is locked? - is word locked or the value locked??) is this fn locked?? then we could replace add word directly with builtin object. 
	1 lookup less. Same for function object and maybe for tuples too?
	would we want a function bind to some external value bind to the word of the value or value itself??? What would these two options mean in concrete scenarios.
	What makes more sense. What is more stabile, less error prone, more functional, immutable ... ???

We made first primitives and we can run them via objects (not string code).

## 24.01.2019

Next we make builtins evaluate via string code. 

We also need to figure out what EvalBlock() takes and returns. Could it return just object? Since it had it's own series as code and in some times it's own env.state.

When evalBlock is a function it has it's own clean state. When we eval a block via DO or similar it doesn't. Special whitelabel functions just take specific words.

Should we split EvalState which includes env, pointer, series, result. While env holds words1 words2 wordsn state. Maybe state should be separate because we have many states.

Think about it ... we can also refactor this later ... 

## 26.01.2019

Ok. So here we go. EvalState now holds two things. 
 
 * State of current execution: current series, pointer to current execution position, and optionally return object
 * Environment which again holds
  * words
   * words1 - indexes to words (array)
   * words2 - words to indexes (map)
   * wordsn - number of words (int)
  * state - word indexes to values (objects)
  * parent - parent env ... hm .. wordsindex should be just one so it doesn't make sense to link to parent env here .. or to have more than one envs (with words in them)
    - separate them
	
So now it should be:

 * ProgramState
  * series  	([]Object)
  * pos  	(int)
  * ret 		(Object)
  * env		(Env)
   * state   (map[int]Object)
   * parent 	(Envi)
  * idxs		(Indexes)
   * words1  ([]string)
   * words2  (map[string]int)
   * wordsn  (int)

Cases when we create EvalState
 
 * load a script
  * idxs (words*) are set by loader itself
  * env is created empty and is set by evaluation of script (subenvs are created when needed), parent is set to current env
  * series is set to current series, pos to 0, ret is initially nil

 * enter a user function
  * idxs are the same and don't change by the script
  * new env is created and populated by arguments or whitelabel words
  * series is set to body of the script and pos to 0, ret nil

 * exit a user function
  * env before entry is set back to env
  * old series is retrieved bac and position to where arguments ended
  * last expression's result is set to ret

 * enter a do block
  * env stays the same
  * series is backed up and set to the one in block

 * enter the let block
  * new env is created populated by let values parent is parent env (could we stay in same series and flag the new ones as the ones to be cleaned afterwards? - would it be faster even)
  * series is set to series in block, pos to 0
  * last expression result it set to ret

Let's set this all up now and make it work. About EvalBlock, because block can change the Env we enter and return it also. So we probably passwhole ProgramState

## 27.1.2019

We made builtins work yesterday. Then in matter of minutes we added functions for ADD, INC, IF, EITHER, and today LOOP. Since we have loop we were able to do first performance tests.
The bad news is that Go version is only slightly slower than JS version. 

loop 10000000 [ ... ] are aproximately
2s for 1 variale lookup
3s for 2 variable lookups
2s for 2 setwords (faster than JS)
12s form two function calls to add 

 * Word lookup is similar to JS. If we remove a check to parent (even if variable exists in current scope) scope it's almost half faster, which is strange since there is only one IF
 * The builtin seems the slowest here so we looked into code and experimented:
  * it seems the builtin function itslf (with many checks etc) has almost no effect on speed. We returned env.Integer{123} directly and even this didn't improve speed
  * if not this then there could only be word lookup, or setup of calling the function itself ?
  * IDEAS TO TEST:
   * we could reduce word lookup by changing word with builtin itself in the series
   * we could test the function setup by instead od calling the function returning value itself there
   * maybe we are making some value copying that is unnecesary when calling a builtin
  * Looking at CallBuilting .. it looks like argument evaluation could be taking a lot of time. We remove the call to function itself and just eval arguments but IT TAKES 0.00s with
    10000000 repetitions - STRANGE .. So it must be the function calling. What if its because of variable argument count?? I suspect regular cuntions in Go can't be that slow... :/
	--- not strange ... without function calling loop doesn't even "loop"
  * **yes argument collecting is taking us time** .. if I do it twice instead of once time doubles!!
  * **and function calling**

a) loop 1000000 [ oneone oneone ]
b) loop 1000000 [ add 1 2 add 1 2 ]

Normal:
a) 0.33  b) 1
Two arg collecting:
a) 1 b) 2.5
Two go function calls:
a) 2.7 b) 5

So both take considerable time. Function calls even more. Now let's see if add returns integer immediatelly ... if function body has effect.
With immediate return of function add
Normal: 		0.45 	1.05
Two arg: 	0.86		2.17
Two calls:	2.38		4.61

SO THE BUILTIN FUNCTION CONTENTS DOESN'T AFFECT IT MUCH ... it's the call itself .. maybe call with static args would be faster? 

1) make function redux which compiles in the builtins instead of words in blocks and loop uses that... see where it leads us this
2) try things to make evalExpr faster. thing about where should be pointers to things and where not. there is plenty of perf tools: https://github.com/golang/go/wiki/Performance
3) try calling a function with static number of arguments and see if it's faster

### Pprof

Non-variadic builtin function call and removal of allocating arg array.

I started using pprof and generated pdf, looked at it ... then opended pproff and looked at top5 and then listed top function. This showed me that the 
variadic function call to builtin function takes the most of time and also allocating array for arguments. I tried the static arguments on function and
also removed the need to create array. Builtin function calls will now be limited to 5 arguments via this pattern.

This improved speed of function call in loop by more than a factor of two!! We are still slower than rebol though. But I see that there is tons of tools
in Go ecosystem that will enable us to speed things up naturally. In general allocations seem to be costly, so we should focus on them further.

I noticed thr ps.env.Get call takes considerable time to. Maybe some optimisation here ... caching. Somewhere was mentioned that calls on interface methods 
can be absurdly slower than calls on concrete methods. We should explore that. Maybe change our internal representation of objects alltogether.

How did we use the tool:

added one line to Main()

/usr/local/go/bin$ sudo ./go tool pprof --text ~/go/src/Rejy_go_v1/Rejy_go_v1 /tmp/profile994276150/cpu.pprof > ~/go/src/Rejy_go_v1/pprof3.txt
/usr/local/go/bin$ sudo ./go tool pprof --pdf ~/go/src/Rejy_go_v1/Rejy_go_v1 /tmp/profile994276150/cpu.pprof > ~/go/src/Rejy_go_v1/pprof3.pdf

/usr/local/go/bin$ sudo ./go tool pprof ~/go/src/Rejy_go_v1/Rejy_go_v1 /tmp/profile658246329/cpu.pprof
(pprof) top5 
(pprof) top5 -cum
(pprof) list evaldo.EvalBlock ... shows exact lines in function and how much time they take

## 28.01.2019

Today I read more about performance coding in go. Allocations, stack vs heap (don't use pointers unless you need to) ... avoid Interfaces in hot code. Escape analysis ...

* Avoid allocations
* Avoid runtime info (sizes)
* Only use pointers where you need to (changes to obj)

I then removed the interface, which didn't make much change. Then I changed *Object at some hot function and it seed up by 30-40 % !!

Now we are in the ballpark of Rebol!

a) loop 10000000 [ oneone ]
b) loop 10000000 [ add 1 2 ]

a) 0.9s b) 1.2s

This is awesome! Rebol b is around 1s

a) loop 1000000 [ oneone oneone ]
b) loop 1000000 [ add 1 2 add 1 2 ]

Normal:
a) 0.33  b) 1 		PREVIOUS 
a) 0.20  b) 0.35 	NOW

two add 1 2 calls in the 10.000.000 loops took at first 12s not take less than 4s. 3x speed-up!!!	

## 03.02.2019

### user function

So I am making user functions. Let's first make execution of function object. This helps us define the function object. Then we create the function making function.

Hm ... I made the frist version of user function and a test with user function that just returns integer. The WEIRD thing is that in loop it runs wery strangely fast :/ ...
So fast that it's very suspicious. But loop does behave lineary to number of loops.

{ loop 1000000000 { fun1 } } ... one billion loops just takes 12s? Our normal 10m takes 0.12s

fun1 is defined as:
		body := []env.Object{env.Integer{2345}}
		spec := []env.Objec
		*env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), *env.NewBlock(*env.NewTSeries(body))
		
let's make a function that calls a builtin. Then let's make a function that takes one argument in tests. Since it's so fast I suspect we have some bug around env.

We had an weird error that only showed when calling user func inside a loop. Took me 1h ... at the end it was we didn't record the Series after evaling all args so
the arg got evaled again and it was return of the block. While solving this I saw what looked like an easy way to do recur (as in clojure).

I fixed the error but now the example above runs 100 times slower :( and I can't seem to find why before it ran fast. It now seems even the arg evaluation on no args 
takes lofunction even w/o funct so it doesn't make much sense why it ran as fast as it did before. We will try to figure out this when we optimise in general.

If we always accept and return ProgramState it could also be passed by value not reference, which showed to be faster in general. We should try it in specific git branch later.

### recur

Recur would maybe just need to overwrite arg words with arguments to recur, reset the current block series pos. So it could be user or builtin function. We could make user functions
that have the access to current programstate or make it accessible via flag.

Test if we can do recur similar to clojure one. Since functions in rejy are of fixed arity we would need recur1 recur2 recur3 and recur [ ] which is less optimal
otherwise word recur could somehow be bound to correct version or args depending on number of args of func. Try this at first.

we got the recur working. But it can only recur on top level of function ... not inside any block for now. To make it recur inside a block we would need
to return a recurobject that holds arguments and when toplevel gets it it does as we do now. To either way test it we created recur1-if which takes a condition 
too.

factorial: fn  { n a } {
	recur2if greater n 0 subtract n 1 multiply n a	
}

later when we have better parser and opwords we could write

factorial: fn [ n a ] {
	recur2-if n > 0  n - 1  n * a
}

Skušajmo naredit zgornji recur. Dodal sem subtract, multiply, recur2if ... to bi pisal v kodi, tako da bom dodal še fn builtin, da lahko naredim funkcijo.

### fn

Najprej preizkusimo dodani fn builtin. Potem naredimo factorial z recur.

recur2if recur1if in recur3if sem dodal. Recur2if se je potreboval pri factorial, recur3if pri fibonacci. Oba sta obcutno pohitrila izvajanje. Fibonacci ocitno,
ker je algoritem veliko bolj optimalen, ker je bil tudi fib(50) takoj izracunan. factorial kjer mislim, da je algo podoben, enako št. rekutzij je pohitril x2.

### performance compared to JS version

I retested where we are and we are much faster except for some reason in 100000 loops over factorial 12. Fibonacci is 2x faster, loops are > 2x faster too.
Some even 4x. When we will have function inlining it should be even faster. We will also make a branch where we try to make programstate passed by value, to 
see what that does to perf. Contrary to c, such values could be via escape analysis be made in stack, which is much faster than heap. 

Some perf numbers:

USERFN: loop 10M { add 1 2 }  JS: 7,0s  Go: 1.6s
BUILTIN: loop 10M { add1 1 2 } JS: 20s   Go: 11s

USERFN: fac: ... loop 100k { factorial } JS: 14s Go: 4.2s
USERFN: fib .... fib 30		 			JS: 20s Go: 8.7s

So our current Go version is aprox 60% faster than JS Rejy. I hope the opwords in evaluator won't make it any slower.

Next thing would be, to add the opwords support and see if it slows the evaluator. 

I added it to git today also. A separate branch should be done where we test passing programmstate by value in all cases. Maybe also all values in in 
it should be values. So we see if Go can make this faster as it did with some other changes to val vs ref.

# 24.2. Implementing simple strings, first peg, then object, then some builtins. Implemented, basic test done.

# 7.4 Implementing builtins for blocks and adding tests. nth, peek, Next

one interesting observation: pop doesn't make sense as a function right now can just return changed object, but can't return something else and 
change object passed to it (as pass by reference). This is good in view of preventing side effects. We will see if we would need this reversed.

another interesting observation: when testing I saw the practical difference between . and | 

{ a: { 101 102 103 } b: a .nth 1 |add 100 }" // 202 -- returns second value in block and adds 100 to it
{ a: { 101 102 103 } b: a .nth 1 .add 100 }" // error -- tries to return 101th block value (adds 1 and 100 and returns it as arg to nth))

since the forward direction of code is noticable feature of this lang ... maybe name should be *fwd lang* or something like it. awk would then be fwk

NEXT: things to do would be to implement handling of errors (so that if they are meet at any stage they get handeled specifically, not just returned)
NEXT: thing to figure out would be to make all words basically first argument type/tag sensitive (the short word goal). We need to dispatch on primitive values
and custom objects that could be some kind of structs / tuples / object (that will later get validation directives) or various native/binary object like canvas opengl handlers etc.

Basically, we need to figure out what these objects/tuples in rye or fwd will be. They should be as lightweight as possible. Maybe with some copy on write optimisations and able to belong 
to one or more classes / kinds. Kind is defined by validation rules. We can enforce the tuple to kind and we get kindered-tuple out of it or nil (or validation errors object).

THINK: We want some inline setwords option too that works in combination with op/pipe words. 
THINK: We also need to figure out what to do with regular operators. Are they just op/pipe word types?? + * ?

# 14.4.2019

made another branch kinds

core kinds have predefined integers as in IntegerType, BlockType, etc

# 02.06.2019

added getKind that returns integer of type for native values. Now we should register kinds so that they have these numbers ...
can we call it first so it will take those indexes?

We added register function and generic builtin. The map accepts objects now. 

Now we should dispatch on first argument, but if all are just words how would we know to lookup in Gen before ve evaluate first arg
if it's firs arg at all. So we need to make a distinction Word are local words, word is a generic word (linked to function or builtin)

But for now, so the test will pass whil we implement the change we make it reverse. Word is generic. Since all our builtins etc are now not
generic yet.

So we make another type of word Genword (that for now starts with a uppercase letter)

Made it work. It was quite simple to do. Just added evalGenword with few different lines.

After that I also made generic 'type 'Add ?add work with builtins. It basically already worked, I just needed a getword type
whis is so far ?word. because :word will maybe be used for left setword.

It seems we will need to make pipewords and opwords genword variants too ... but it shouldn't be complex and better to separate this 
at load time, not interpret time.

WHAT TO THINK ABOUT NEXT: will all builtin functions be generic functions? Probably makes sense, so we can in context reuse 
the words etc.

make the tuples / objects object. They are word/object dictionaries with validation rules ... each object has a kind determined
and you can define generic functions on those kinds. Binary objects also have kinds ... like image / connection / ...


// 27.9.2019 

Started adding SPRUCE LANG SUPPORT.

I started making a spruce/build parser. It's a indentation aware tree representation parser.
I created a SprNode ... a node in a tree that holds a Rye Object (word, string, block), list of child SprNodes, depth
and a parent SprNode.
I added a function to find the right parent to add a node to based on a last parent and depth.

Next. I should add the Rye loader, that turns strings to Rye Objects (word, string, block). And make a unit test that parses
fist tree.

when
 email "help string"
  is 
   received 
    do
     {todo:block}
     [ on-event 'email-received { 'it 'email } todo ]
   sent
    do
     {todo:block}
     [ ]
   blocked
    do
     {todo:block}
     [ ]     
if
 {cond:boolean}
  do
   {todo:block}
   [ core/if cond todo ]
it's
 {key:string}
 [ core/at it key ]
{a:boolean}
 and
  {b:boolean}
  [ core/and a b ]
{val:string}
 includes
  text
   {text:string}
 is
  from
   domain
    {domain:string}
	
I also added the tests file to loader and it basically works.

NEXT: given that we have a basic parser and with it a basic tree, we can make a simple walker / executor next.
think about making proposer too and how we could add tree branches to existing tree ... etc

how could we make a simple console based IDE ... or emacs based IDE? How do IDE-s connect to language runtimes at all.

//29.9.2019 #SPRUCE 2 

Ok, so basic expression works. We would next need to add argument definition and pickup, but to do it as example above we
would need to extend the parser with these forms {name:type}. Are there any other forms that we already parse for?

example option #1:

add:
 numbers
  {a:number}
   and
    {b:number}
     { add a b }
	
Sum: add numbers 10 and 5 
. returns 15

could we remove the {}, since they are already used for blocks too?

add
 nubmers
  first:number
   and
    second:number
	 { add a b }
	
Not enough visual distinction so looks cleaner, but the tree looks more complex, because
args are not distincted apart imediatelly. We also might need to pack in more logic in this
pickup words, so just : is not distincion enough. We might pick up multiple types, add transformers, 
composers etc, so a whole mini language might get in later.

my:
 number
  { 041741612 }

in:
 high
  alert
   mode
    { get-alert-mode .equal? "H" }
	
when:
 pinged
  if
   {cond:boolean}
    do
	 {code:block}
	  { if cond { do/spr code }}
	
time:
 is
  between
   {t1:time}
    and
	 {t2:time}
	  { get-time .between t1 t2 }
		
when pinged if in high alert mode and time is between 20:30 and 06:30 do {
	sms my number with text "something weird is going on"
}

alert:
 numbers 
  { { 041741612 041741623 } }

for:
 all
  {list:block}
   do
    {code:block}
	 { for-all 'it list { do/spr bind [ 'It it ] code } }

when pinged if in high alert mode and time is between 20:30 and 06:30 do {
	for all alert numbers do { 
		sms It with text "something weird is going on"
	}
}

...

So, first we need to add those argument placeholders {name:type} to Rye parser. 
Then we can use them inside Spruce builer and user quite easily.

Then we need to decide, where the varables picked up get stored and for how 
long. It seems obvious that each start of spruce expression should have it's own
namespace, and namespace is used in code, if there is one. After the end of 
the branch that namespace is discarded.

There could later show to be some special forms to directly set It, or decompose
stuff. Or compose to the It object for example.

##composition of object example?

create:
 person
  {It:object:create}
  with *
   name
    {name:string->It}
	 *
   surname
    {surname:string->It}
	 *
   age
    {age:number->It}
	 *

general object could be a Rye tuple instead, that has defined validation rules, and those
are checked and returned as error if not meet when used



create person with name "Janko"
. returns Validation error, tuple Person { surname: required }

## decomposition of tuple example?

get:
 age
  of
   {p:person}
    { at p 'age }
	
or potential automatic decomposition

get
 age
  of
   {p:person:'age->}

// 

Ok so we now have a parser for Argwords. We should make evaluator too now.

First question is where do we store the state. By looking at it I think we should use the ProgramState used in Rye, since we have the same 
Series of code, Env (state / variables), Idxs ... It's basically all the same at evaluation time. We also pass the state to Rye code, so 
it will just use the same ProgramState. This means that we at the begining of evaluation create ProgramState, and pass it to evaluator like in Rye code.

We also need to turn EvaluateBlock to EvaluateExpression, so it can evaluate multiple expressions (and later subexpressions)

## IDEA ... BIG AIML LIKE RECURSICE RULES

when thinkgin about natural interface to our project webapp, the whole AIML spiel I got thinking about it's recursive rules. That is big concept .. biggest 
of aiml and we need to add it too. 

For one, it will make Spruce usable also in processing natural varied user inputs, like aiml but with arguments and more.

To directly make natural text input interfaces similar to workonomic's which needed pyaiml and all sorts of stuff ...

And it might bring interesting, not thought out yet features to "coding Spruce". For one aliases and shorthands, but maybe with some composition etc more?

add..
 two
  numbers
   (* add numbers )

add..
 { sum: 0 }
 numbers <loop>
  {a:number} 
  { sum: sum + a , recur-to <loop>  } 

... does aimls also recur in the middle of the branch. Do we want / need this too? Think of cases ... for now just keep the general idea.

# SPRINT 2019 WINTER (14.11.2019)

Stop adding language features / gadgets (maybe remove some, that aren't 100%). Make existing language work, think about scopes, implement the
generic words, make them work. Then return and make this basic language work. Add natives, try embedding a simplest form of web-app programming.

First we run all the tests.

When tests pass. We look if we would comments out some more obscure features.

Then we look if the generic words work, or how to make them work. What about the uppercase/lowcase method of distincting generic and local words. 
I am still not 100% sure of that and it will break the similarity to rebol.

So, main dillema is, if we make the generic words the default, then we at the same time make Rye quite different to rebol. If the goal is to just make
practical rebol + some other options in Golang, then the generic words aren't default or at least need some other (more noisy than lowercase) 
notation to recognise. 
Login-user: fn [ Id Pwd ]
Sum: read add Name: get-name-from-id Id

The benefit is that local variables which we don't want to have too many sort of popup.

Maybe then it makes sense that GLOBAL variables or (like) constants etc. would need to be all upper case. They could be stored to Global state which is 
more introspective, or even has state management, like reactive state libraries, ...

But if we have generic (kind sensitive functions), then we need to decide how do we assign kinds to values. Primitive values have automatic kinds, 
objects, binary blobs, we can assign kinds to them? Does it make sense ... can one value be of multiple kinds or can kinds be like unions etc?

// OFFSHOOT IDEA ... something like refinements ... (ref) .. also for denoting number of args ... continuation type ..

!person is-union-of (3) !female !male !other ... could we dispatch on arrity too?

join (4) (space) "hi" "my" "name" "is"

... is of type continuation 

join: fn [ !string a b ... ] {
	rejoin join [ a b ] ...
}

join: fn [ !block a b ... (block) ]
	
]

join (block) (map-before) [ "janko" "metelko" "grosuplje" ] [ uppercase ]

// END OFFSHOOT FOR NOW

## HOW DO WE TAG KINDS, TRANSLATE BETWEEN THEN VIA RULES (dialect) OR CODE

If each value could have multiple kinds, then ways to store and lookup them are costly. If there could be type hiearchies, storing is simpler, but traversing
is costly too probably. Maybe there are just direct type lookups, but you can change kinds based on a hierarchy ... or forget the hierarchy .. we just have 
the "p2p" rules, that could convert anything if specified. Also similar types like integer to age and person to employee for example.

For starters each value has just one kind and kind is singular ... we do make translators yes. 

Transators should have their own namespace (and possibly dialect later) ... the type is the word ... but it should be of some specific word type.

add 100 integer< Buyer .get 'age

add 100 Buyer .get 'age >integer

Buyer .get 'age >integer .add 100

Buyer: person< { name "Janko" age 40 } ; tags it with !person 

def-translator !person !age { get _1 'age }
def-translator !person !info { get-all block< _1 [ 'name 'surname 'age ] .join (spaces) }

translator !block !person { } ; just tags the block

print <info Buyer
Buyer >info .print

#tagging a value or translating 

We need option to distinct when we just want to tag some value or when we expect that a translator with some rules, etc will be called
Since we are in this space we can use >type and >Type , type< and Type< for this.

## WE FORGOT ABOUT TUPLE RULES (CREATION VALIDATION DIALECT)

Is the validation dialect called automatically on new block? Probably should be. Or we could have a option to validate it just when we need to.

So when we tag block with tag !tag:

Buyer: Person< { name "Janko" age 40 }

code first uses a translator from !block to !person. If there is none, error happens since Person< is uppercase.
Then on returned data the validator is ran if there is one. Then the block is tagged with "person" tag.

def-validator !person { name: required and string surname: optional and string birth: optional 0 and integer and check ( _1 > 1900 .if { "too old" } ) }


Age: integer< Buyer .get 'age

## So concrete .. hopefully we can as I am not totally up to speed of the state of code

* Run tests to see where we are at all
* Rye objects have a setKind and getKind, kind is is index of a word in indexer, add this
* Each object has just one kind. Native values have hardcoded kind returns (words get preindexed), add native objects getKind
* look where we have the generic words ... can we already create and interpret them?

## Around 1.12.2019 

I made a shell, quite nice on first look and also a simple web server based on golang's echo.

## 15.12.2019 Webserver fwd ...

I could try making webserver script accept request parameters for next step. This way we would also see how we could hangle this in best way.

Echo server does this:

func(c echo.Context) error {
	name := c.FormValue("name")
	return c.String(http.StatusOK, name)
}

func(c echo.Context) error {
	name := c.QueryParam("name")
	return c.String(http.StatusOK, name)
})

We could make it work like this:

<% name: form-val? ctx 'name %>

<% city: query-val? ctx 'city %>

## Next time add query-vals and form-vals ... make out function work

<% probe query-vals? ctx %>

<% 
	data: validate query-vals? ctx { name: required one-word age: optional 0 integer } 
	out get data 'name
	
	out data ? 'name
	
	data ? 'name .out .wrap 'b 
%>

## Next add loader support for ( ) so we can add the Dgraph dialect loading ... something in line of ... (but without the scope/context)

friends: scope {
	
 find-friends: webfn "returns true if two persons by ID's are friends"
 { person1: required integer } 
 { 
	totalRoles ( func: uid ( ?person1 ) , orderasc: val ( roles ) ) {
    		name@en
    		numRoles : val ( roles )
    }
 }

}

OR maybe first should be sqlite ... without parens for now like

bills: scope {
	
 get-finalized: webfn/user-sqlite 
 "returns finalized bills"
 { 
 	finalized: required integer
	number: optional 50 , integer 
	from-date: optional nil , iso-date , check ( this .year .greater? 2001 ) "date must be above 2001"
	to-date: optional nil , iso-date , check ( this .year .lesser? now .year? ) "date can't be in future"
 } 
 { 
	select * 
		from bill 
		where 
			finalized = ?finalized 
			{ from-date ; and date_created > ?from-date }
			{ to-date   ; and date_created < ?to-date }
		limit ?number  
 }

}

# this above seems to be THE GOAL ... the for now ideal declaration of what should happen on given api call

## Some other ideas I had yeasterday 

# json / block collecting dialect inspired to graphql ...

collect data { 
	name: convert ( this .to-upper-case )
	date: convert ( this .to-iso-date )
	friends { limit 10 name: city: }
}

# idea about fwd code flow ... get some object, parse it to discrete variables, use those variables
# like det user object, parse out id, email and username , update event to log using id, email to email with content of username

# get just id, use it
get-current-user ? 'id |log-user-event "signin" now

# get in and email, use them both in one action ... with isn't like idea for implicit ... it's just that when the block is started the current value is already 
# set to the first argument of skip-with ... maybe all skips should have this ... there is no downside I think ... (3)
get-current-user .skip { ? 'email :email } ? 'id |log-user-event "signin" now email

# ? -- means lookup

# get in and email, use them both in separate actions
get-current-user .skip-two 
 { ? 'id |log-user-event "singin" now this ? 'email }
 { ? 'email |notify-user "signin attempt" }


# get in and email, use them both in separate actions
get-current-user .skip-two
 { ? 'id |log-user-event "singin" now this ? 'email }
 { :email if failed { notify-user email "signin attempt" } }
but this could just be one skip ... unless we need that object further, 
the negative is that we loose the symetry between the first and second action, which are equal otherwise

get-current-user
 .skip { ? 'id |log-user-event "singin" now this ? 'email }
 :email if failed { notify-user email "signin attempt" }


# would this return the resul of last block, or the incoming object? based od (3) ... skip-with would just be skip. And the word determines what is returned
after ececution of block ... skip's point is to return the incoming object, but we could have something like cascade, which returns the last object, but this is 
then just normal concatenation with pipe-words I think :) even better if it is .. explore more

# implicit that set's the first argument for all function calls in block ( ) is escape hatch where implicit doesn't work

with ctx {
	name: query? 'name
	num: query? 'num
	id: sesion? 'id_user
	set-session 'user ( get-email-of id )
}

# with is more natural than implicit ...

# Defer - idea from golang 


query-user: fn { id } {
	f: open %file
	defer { close f }
	id: with-sqlite %db {
		select * from user where id = ?id
	}
	write f id
}

### Next ... more sqlite funcs, first graphql 

How would dgraph examples look here:

#go 

q := `query all($a: string) {
    all(func: eq(name, $a)) {
      name
    }
  }`

res, err := txn.QueryWithVars(ctx, q, map[string]string{"$a": "Alice"})
fmt.Printf("%s\n", res.Json)

#Rye
--------------------------------------
api-fn 'users 'find-user ( auth )
{
	name: required string few-words
}
dg: open dgraph://localhost:9080
query dg {
	all ( func: eq ( name , ?name ) ) {
		name
		age
		friends {
			name
		}
	}
} |to-json |echo
--------------------------------------


# validation has multiple functionalities basically

validation dialect can also construct / deconstruct / reconstruct
pick from local values to block, block to block, block to local (block is not really a block but a map/object/tuple/context)

collect { name: :firstname age: integer :age }
set forms? { name: required short-string :pers-name }
map forms? { ... }

# how would we make a first api function
# we need something like context

posts: context {
	paginate: { limit ?page * ?per-page, (?page + 1) * ?per-page }
	table: 'posts
	per-page: 30
	
	get-all: webfn/sql
		"gets all posts, paginated"
		{ page: optional 0 integer }
		
		{
			select * 
			from <table>
			<paginate>
		}
	}
}

# Env is context is object .... can we lock whole object's contexts? ... if we add a flag to it ... what do we get ... 
# we would also need a validated flag probably

ok. So context is the same as env and same as object ... basically execution environment is a first class Rye type.

Then we could have function in-context, which places given context as primary and normal context as it's parent?

do with-context 'gtk {
	init 
	w: new-window
	l: new-label
}

what if we use factor like convention for new ... it clashes a little that words with op as first arg are op-words?

do with-context 'gtk {
	init 
	w: <window>
	l: <label>
}

#idea ... flag for 


# 04.01.2020

Ok so we make the first validation dialect. Validation dialect will not be a pattern / code flow of Rye functions ...

The main reason to not explore this direction is that it's integral part of the language, also used internally, not just to
validate external input, so it must be as fast and light as possible and we want to avoid any namespace lookups if possible. So at least core
validation rules should be executed directly in GO, with no steps to rye environments. 

First validation rules are:

required , optional X , string , integer , number , email

Then we add the adhoc rules for: 
  birth: optional _ opt-iso-date
  email: required email check 'email-taken { .find-user-by-email .found? }
  exid: calc { .add 100000 } ; takes current value and calculates new one
  email-hash: generate { .get 'email |hash }

There will also be a rule for executing and arbitrary Rye function in multiple modalities 
(accepts validation-state returns validation-state, returns true/false, etc)... later

So, firstly ... we implement the dialect words above. For reasons mentioned this dialect should be compact hardcoded function switching on
these words

# Error handling ...

1. Errors print out and or log and stop execution

2. Failures can be caught in code, if uncaught they turn to errors

Error is like early return / unwind of our whole evaluation. How do we do it? Let's make a minimal example and think about it:

let's first make a return work ...

{ print1 return 2 print 3 }
// should print 1 and return 2, but not print 3
return has to pass over multiple scopes ... nested blocks like if {} do {} foreach ...

this can be done via some special return type, that when meet causes exiting etc ... but where exactly do we detect it? ... and checking constantly is costly ... 
could there just be a return flag in current Env that is set and then unset on exiting? And return would be a builtin that set's this flag?


{ print 1 return err 2 print 3 }

should print 1 and then exit. err is the builtin that returns the error object.

{ print 1 do { print 2 err 3 print 4 } print 5 }

Can our code throw error? if error exits immediatelly? Does it even make sense? When do we detect the error? On return of call? It's related to return ... error is like return error


# 2020-03-22

HOW DO WE MAKE A FIRST WEB API 

Added postgres binding. If we are trying to solidify thing there is tons of things to do. So while developing new, we should also solidify things like Errors for builtins.

- ProgramState should have a error function that returns the error and sets the flag.

Spreadsheet should have some basic functionality:
+ list of columns
+ count of columns
+ count of rows
+ to html
+ to text
- to json
. get column out as a block of values
. get row out as a blok of values
. get row as raw-map out
. get sum, max, min, custom reduction of a column
