# Planning
---
## For 0.0.16 [+]

### Subc flag default [++]

Current default is that builtin words load into current context when Rye starts. `--subc` flag starts Rye shell or file in a subcontext, meaning that all builtins are loaded in the parent context.
Positive now is that with `ls` you can quickly see all builtins you have, negative that your words are lost between those.

We will make subc waythe default one, so `ls` will only show your context which will be way clearer. We will add function `lsp - list parent` which will show in this case the builtin. Benefit is also that
you then can't change or owerwrite parent functions (but you can owershadow them - which is better than first - think of some warning about this).

### Evaluated blocks [++] @Rok

I left the difference between { } and [ ] open for years. Currently both work the same, but we use { } . One benefit of mixing would be that if you alternate them it's easier to see the matching parenthesis sometimes. But it's 
a dubious one - as it would also confuse new viewers of code and colors can do the same. I was making Fyne integration. There is a constructor for container and it accepts any number of widgets. The best way to do 
this in Rye was to put them into a block, but blocks aren't evaluated by default in Rye (Rebol). So you need `vals` (reduce) in front and there is an extra step where runtime costructs a block and then a funtion evals (reduces) it.

If [ ] evaluate by default we reduce the need for vals function (which I couldn't find a good name yet btw. Rebol's reduce is taken, eval would be the most correct but would strongly conotate javascript's eval which does a different (bad) thing). 
So some code becomes cleaner and runtime handles these cases in one step instead of two, which is not unimportant as it can be used a lot.

### Builtins inside context's [++]

At integrating Fyne also it was decided that builtins should have and option to be loaded in their own context. Once a native is created we use generic methods for namespacing, but to construct them words need to be namespaced. So we have either 
fyne-window, fyne-button, fyne-entry, ... of we put all those in it's own fyne context and we can have fyne/window, fyne/button, fyne/entry of just execute code inside that context and have only window, button, entry ... This is the best way for 
multiple reasons so builtins must support loading in their context. If you would want at build time determine what builtins should load in their own context and what directly, or if it should be decided by the builtin maker is still an open question.

Too much variability can then complicate reuse ... this would need to be determined by flags and build flags are already a per-project parameter.

For now it's statically determined at development of builtins. We will explore more dynamic options later.

### Rye shell / console naming [+-]

I need to decide what Rye REPL will be called. REPL is not user friendly word. I used shell, but it can get confusing with Bash, Zsh, maybe one time Ryesh :P. I use console here and there so it's not unified. Google gemini suggests Rye console 
is approachable than shell, which is contrary to what I thought.

### Web shell - multiline [+-]

We shell needs to support multiline, for some examples in docs to work. It can work the same as Regular shell, where if the last character is space it goes into multiline mode. In this case the space character should be made visible, via 
something like middot or a newline like character. Or / and it could look for open parenthesis and go into multiple lines until it's closed.

---

## For 0.0.17 [+]

### Command line arguments improvement [++]

* add --do option - does rye code after it loads files (named, dot (.), or cont mode)
* add cont command - continues from previous shell save
* Redo flag handling to flag library, include help

#TODO: add more examples:
```
rye cont --do 'loop 3 { init }'      # runs last saved state and then --do so it's more visible and explicit you are about to run something ... CLI injection? 
rye . --do 'init'                    # runs main.rye and function init
```

### Table improvements [+++]

More where options:
* where-match regexp
* where-contains
* where-in
* save\csv
  
**Group-by** function creates a new table with category column and agregated columns:
```
spr .group-by 'category { price: avg  qty: sum  category: count }
```

**left-join** and **inner-join** joins two table given two columns of them.  
```
users .left-join groups 'group_id 'id
```
### Math dialect improvements [+++]

* added support for custom functions, unary or binary
* function arguments are defines inside parenthesis, like math syntax usually is
* added math subcontext that includes typical math functions and math dialect, added first few

---

## For 0.0.18

### Rye force console, silent, do, help, dot behaviour [++]

Run Rye by running a file or . and if this option is set still go to console after ending file. Like cont does, but for any file. Command is named console. Add proper Arg parser and help.

If the file argument is . or some/path/. look for main.rye in that location and run it. 

### Cont cli command - Save\current\secure [++]

Let state saves be encrypted with password and console_.....rye.enc ask for same password. Demo it with simple contextplay password manager.

### Additional functions, value types [++]

+ walk - useful for dialects, recursive algos, ...
+ xword, exword improvements, xword accepts args, equality - matching still works
  
### Import function , current path / script [++]

Import function like "do load %file" but only looks for files local to the main script file or current script file basically. 
Try to make import always relative to current script, so interpreter should have a concept of "current script" at loplevel. Maybe a specific "do" variant that is invoked in improt and is script location aware.

This would also be needed for loader errors in case of multiple files, so you know what file the loader failed.

---

## For 0.0.19

### Improve errors with current path / script [+-]

At least loader errors should be able to display filename next to location. Toplevel code at least should also be aware of script filename, or should we look in direction of a stacktrace. Would we need to manually manage the stack of callees'
maybe only in debug flag since it would impact performance. Maybe functionname => number of calls so recursion doesn't push out the stack info.

### Liner with standard ansi colors, some improvements [+-]

Test using standard colors, we will see if they work in Emacs ansi-term then, we will see if maybe general terminal theming works on them, also xterm.js probably has theming, test it, make build flag if possible, update refaktor/liner for it.

We could also try to add some low hanging fruit improvements of syntax highligher for specific word types and jumping logic when we are already updating this.

### Web console - fwd [--]

- Make paste into shell work [+]
- More key combinations (ctrl-l, ctrl-d, ...) [+--] 
- Sole keys (pageup, pagedown) [--]

### Math dialect fwd

Let math dialect call rye functions also, not just builtins. It would be good if we could just use function "math" ... not context math and in it function math or math/do ... but still preserve the usable context too.

Pratik is adding standard math functions from https://pkg.go.dev/math to the context. 

### Devops context fwd [+-]

- Integrate relevant function from https://github.com/shirou/gopsutil [++]
- Create standard commands / utils like cd / ls / mkdir / cp / mv [+-]
- Integrate awesome script library for many standard piping commands https://github.com/bitfield/script [--]

### do_main build flag and Android test [+-]

if build flag do_main is used make the dot behaviour work even without the dot. Usefull for distributing binary and main.rye , also to test to produce a mobil APK with Fyne.

---

## LATER

### Mod-words - experiment

`word::  ::word` would be mod-words and would allow changing existing values to words. So set-words would only create and fail if word already is set in current context. Mod-words would only change and fail if word
is not yet defined. We rarely modify at all ... to much modifying is a smell that code could be written better. word:: visually is not horrible, or that noticable, but jsut noticable enough I think. So let's do this and
we will see what practice shows. For starters just add it to loader but it could behave the same as set-word so we woule volontarily use it and see hot it looks and feels at all. If it seems ok, we change the interpreter.

### MakeArgError improve output

We could try to display more information at MakeArgError. We know what type the value was. We don't but if we knew the names of arguments it would be even better. But builtins in Rye currently don't have named 
arguments at all. And we don't want to store that per-function in runtime, as it would make them heavier. Maybe stored outside somehow? Think about this ... if we would be generating Go code this would also be 
generated. Maybe try to see how can we make it look and if we can retain simplicity and ability to compile and load the Go versions. Maybe we can even better use the go versions as they wouldn't be defined in 
and Array.

### Rye evaluator EvalExpr improvements

Go over the various EvaluateExp functions and solve some confusion around them and .with { x , y } command

### Rye evaluator profiling

Check what pops up of we profile a fibonnaci or a game of life example

### Save and load history

Save / edit with emacs / load back

```
rye cont --do 'loop 3 { init }'      # implemented ... --do so it's more visible and explicit you are about to run something ... CLI injection? 
rye . --do 'init'                    # TODO runs main.rye and function init
rye --do 'print 123 + 123' --quit    # TODO
```

---

## RANDOM

### ~~Smart history v01 (just ideas)~~

Just very rough idea

* the console keeps history
* there could be console command that removes last line from history, and maybe undoes the state if possible (shows history)
* after line is commited, it could be determined if any new values were created and stored which are then used, lines cerating unused values could be purged (it's not always certain if something was used, a gui widget just by existence can be usefull)
* you edit this in context/mode ... like a function mode, context, block, you can define enty parameters arguments for function, and outputs, which can generate tests
* if you are in a function, or any other limited context you can view already added lines and restrt to certain point, edit the lines ... whole interaction / UI for this is still unclear
* How you enter function mode ... just by writing a start of a function or somehow else, how do you set the test arguments? ... does it ask you, do you just use set-words ... do you define them at call time?
* Focus just on functions first and try to make the most natural workflow with as little "breaks" and new behaviours as possible, maybe just mimic coding or using shell as much as possible
  

