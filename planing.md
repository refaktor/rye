# Planning

_Planning in the open. You are welcome to comment or implement the plans :)_

## For 0.0.16

### Subc flag default

Current default is that builtin words load into current context when Rye starts. `--subc` flag starts Rye shell or file in a subcontext, meaning that all builtins are loaded in the parent context.
Positive now is that with `ls` you can quickly see all builtins you have, negative that your words are lost between those.

We will make subc waythe default one, so `ls` will only show your context which will be way clearer. We will add function `lsp - list parent` which will show in this case the builtin. Benefit is also that
you then can't change or owerwrite parent functions (but you can owershadow them - which is better than first - think of some warning about this).

### Evaluated blocks

I left the difference between { } and [ ] open for years. Currently both work the same, but we use { } . One benefit of mixing would be that if you alternate them it's easier to see the matching parenthesis sometimes. But it's 
a dubious one - as it would also confuse new viewers of code and colors can do the same. I was making Fyne integration. There is a constructor for container and it accepts any number of widgets. The best way to do 
this in Rye was to put them into a block, but blocks aren't evaluated by default in Rye (Rebol). So you need `vals` (reduce) in front and there is an extra step where runtime costructs a block and then a funtion evals (reduces) it.

If [ ] evaluate by default we reduce the need for vals function (which I couldn't find a good name yet btw. Rebol's reduce is taken, eval would be the most correct but would strongly conotate javascript's eval which does a different (bad) thing). 
So some code becomes cleaner and runtime handles these cases in one step instead of two, which is not unimportant as it can be used a lot.

### Builtins inside context's

At integrating Fyne also it was decided that builtins should have and option to be loaded in their own context. Once a native is created we use generic methods for namespacing, but to construct them words need to be namespaced. So we have either 
fyne-window, fyne-button, fyne-entry, ... of we put all those in it's own fyne context and we can have fyne/window, fyne/button, fyne/entry of just execute code inside that context and have only window, button, entry ... This is the best way for 
multiple reasons so builtins must support loading in their context. If you would want at build time determine what builtins should load in their own context and what directly, or if it should be decided by the builtin maker is still an open question.

Too much variability can then complicate reuse ... this would need to be determined by flags and build flags are already a per-project parameter.

### Rye shell / console naming

I need to decide what Rye REPL will be called. REPL is not user friendly word. I used shell, but it can get confusing with Bash, Zsh, maybe one time Ryesh :P. I use console here and there so it's not unified. Google gemini suggests Rye console 
is approachable than shell, which is contrary to what I thought.

### Web shell - multiline

We shell needs to support multiline, for some examples in docs to work. It can work the same as Regular shell, where if the last character is space it goes into multiline mode. In this case the space character should be made visible, via 
something like middot or a newline like character. Or / and it could look for open parenthesis and go into multiple lines until it's closed.


