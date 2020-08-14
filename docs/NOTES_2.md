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

```factor
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

```factor
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

```factor
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
