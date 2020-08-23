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

- [ ] add to repl the "loader mode" where it doesn't Do the input but just loads it and shows loaded values
	  this will be helpfull in debugging the more exotic loader tasks. Right now (10.08.2020) Void priority
	  clashes with the word beginning with underscore (which it shouldn't)
- [ ] fix the underscore word - void clash in loader

## Main open design questions

- [ ] exceptions, failures, errors
	[ ] test the idea we had and implemented in various scenarios (faliure if not handeled or returned becomes error, return words)
	[ ] if it makes sense, make it work 100%, I suspect it doesn't work in some pipe-word cases
	[ ] will we need to add try/catch for specific cases
	
- [ ] kinds (generic words? namespaces? validation dialect? validated status? ...)

- [ ] contexts (scoping, extending, isolates ...	)
