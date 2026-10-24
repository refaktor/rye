# Tests

Tests are generated from the Go comments above builtin definitions. `regen` writes three `.info.rye` files, one per source directory: `base.info.rye` (evaldo), `baseio.info.rye` (including IO, commands, OS and HTTP), and `batteries.info.rye`. Each source `builtins_*.go` file is a labelled section within its directory. The same files drive tests and HTML reference pages through `main.rye`; there are no duplicate topic inputs.

## Comment format

Each builtin is documented by the comment immediately above its map key. The
recognised sections are:

```go
// ##### Section Name ##### "One-sentence section description."
//
// Tests:
// equal { expression } expected
// error { expression }
//
// Example:
// ; Short, readable usage example.
//
// Args:
// * value: What the argument means and which types it accepts.
//
// Returns:
// * What the builtin returns, including its type.
//
// Tags: #category #topic
```

`Example:` and `Examples:` are both accepted. `Tags:` is metadata for the
source comment; it is not rendered as an argument or return value. Keep each
structured line on its own `//` line. Blank lines are allowed between sections.

To find out more about the comment docs and format read this: https://ryelang.org/cookbook/improving-rye/one-source/

# Generating .info. files

To generate the info files from builtins, run `./regen` (from `tests`, or `./tests/regen` from the project root). The script builds `cmd/rbit` from source into a temporary binary, then uses it to parse the Go code (Go is required). Edit the ordered file lists in `regen` when documenting additional source files. It validates inputs before replacing output and generates each source only once.

# Running tests

Once you use `regen` you can use the main.rye in this folder. To run it use the dot shortcut: `rye .`, this will show you help information.

```
# To list all tests groups
rye . ls

# Run all the table group tests
rye . test base
rye . test baseio
rye . test batteries

# Runs all the tests
rye . test
```

# Generating function reference

The same tool tests/main.rye also produces the reference docs from .info. files. You can see the docs online:

https://ryelang.org/info

```
# To generate html docs
rye . doc
```

This writes `base.html`, `baseio.html`, and `batteries.html` in `tests/` (the generated HTML is git-ignored). The sidebar links between the three directories, then lists source files, sections, and builtins. Use the sidebar search to filter the outline.
