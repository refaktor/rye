## Current

- [x] start writing intro
- [x] change loader for one character (operator) op-words as in notes (2020-08-08)
- [ ] implement python like for-each without the injected value, because it would be to early in the intro to talk about that
- [x] continue intro page 3
- [x] implement fix-either, fix-else
- [x] add some basic file IO natives
- [x] add returns, collect, collect-keys
- [x] in loader implement filepath type %file.txt that is the same as file://file.txt
- [ ] implement multichar operators in loader like +\ , +_ , etc
- [ ] add additional file IO operations like read write to file-schema
- [ ] try implementing cases control structure
- [x] implement keep combinator
- [ ] implement divides? functions

## Code org

- [ ] rename evaldo module/folder to eval, since it includes all dialects not just do
- [x] move intro and notes to folder ./docs

## Long term

- [x] add to repl the "loader mode" where it doesn't Do the input but just loads it and shows loaded values
	  this will be helpfull in debugging the more exotic loader tasks. Right now (10.08.2020) Void priority
	  clashes with the word beginning with underscore (which it shouldn't). => Don't need this just enter a block of code
- [ ] fix the underscore word - void clash in loader

## Main open design questions

- [ ] exceptions, failures, errors
	[x] test the idea we had and implemented in various scenarios (faliure if not handeled or returned becomes error, return words)
	[ ] if it makes sense, make it work 100%, I suspect it doesn't work in some pipe-word cases
	[ ] will we need to add try/catch for specific cases
	
- [ ] kinds (generic words? namespaces? validation dialect? validated status? ...)

- [ ] contexts (scoping, extending, isolates ...	)

## Kinds (04.10.2020)

- [x] make the Kind object in env
- [x] make the kind builtin that creates the kind object (sets the block as spec)
- [x] make a simple kind constructor, that accepts RawMap, validates it and returns a RawMap with specific kind

## Shopify example (11.10.2020)

### dict, list, json
- [x] rename rawmap to dict in code
- [x] add another object list which is like rawlist (non indexed and non-boxed values. dict and list are what you get from go-s json parse!
- [x] make string loader work with " " and ' '
- [x] try how it looks if dict is direct map without boxing
- [x] make json load to these objects ... how do we do with one in another exactly ... try
  - [x] let json-parse load array to list
  - [x] let json-parse load object to dict
  - [x] let json-parse load array of objects ... figure out if we convert to dicts at accessor or after parse? .. accessor
  - [x] make inspect work better for dict and list and list of dicts
- [x] make map work for lists too, not just blocks
- [ ] make -> work on blocks too
### repl
- [x] try making last value ob line in console be injected block to next line
- [x] improve the distinction between probe and inspect (probe ob object is what you want to print, inspect is more inspective)
- [x] if error happens leave the last result alone
 	