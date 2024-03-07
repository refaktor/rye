# Planning

## For 0.0.16

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

## For 0.0.17

### Web shell - make complete

Try to make behaviour as complete as possible. Paste into shell doesn't work, Some key combinations (ctrl-d, ...) , sole keys (pageup, pagedown). Test and list here:

### ~~Smart history v01 (just ideas)~~

Just very rough idea

* the console keeps history
* there could be console command that removes last line from history, and maybe undoes the state if possible (shows history)
* after line is commited, it could be determined if any new values were created and stored which are then used, lines cerating unused values could be purged (it's not always certain if something was used, a gui widget just by existence can be usefull)
* you edit this in context/mode ... like a function mode, context, block, you can define enty parameters arguments for function, and outputs, which can generate tests
* if you are in a function, or any other limited context you can view already added lines and restrt to certain point, edit the lines ... whole interaction / UI for this is still unclear
* How you enter function mode ... just by writing a start of a function or somehow else, how do you set the test arguments? ... does it ask you, do you just use set-words ... do you define them at call time?
* Focus just on functions first and try to make the most natural workflow with as little "breaks" and new behaviours as possible, maybe just mimic coding or using shell as much as possible
  

