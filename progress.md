# Progress fall 2023 and on

There are still many open details, even some inconsistencies between various ideas, but in the rought the main design of Rye is made and somewhat tested in the wild. Next step is to stabilise the language to make
it potentially usefull. I think that only with concrete use in the real life will the open detailed questions show it's asnwers.  

## Builtin functions

Besides literal data, everything in Rye (including usual reserved words - if, fn, for,...) is a function. So the language itself is and interpreter (in fact multiple interpreters because we have multiple dialects) and a 
collection of builtin functions. To stabilise the the language also means that we have to stabilise the builtin functions. There are core builtin functions (fn, if, join, first, ...) and there are multiple integrations (http server, sqlite, ...), 
all in all there is a lot of builtin functions. 

### Levels

To have some overview over the stage the builtins are in we will define 5 levels. Main builtins are defined in builtins.go all other are is specific files, for example builtins_http.go, builtins_sqlite.go. For each file/integration the level 
can be defined. If the level is not written it's considered lvl0.

  lvl0: builtin is made, works in some example cases
  lvl1: + all docstrings are written, all argument errors are handeled consistently
  lvl2: + argument types are displayed, tests and reference docs are written
  lvl3: + builtins handle all the argument types it would make sense they should handle

## Interpreter

Interpreter hasn't been majorly changed in a while. In general it seems to handle all we've trown at it, but there are some edge cases probably, especially around more obscure ideas in Rye which are not 100% sure to be included. Also more
tests should be made how various features interact with failure handling. This means that we should start writing a set of test that test various facets of the interpreter.
