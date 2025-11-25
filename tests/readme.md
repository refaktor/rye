# Tests

Tests are generated from the Go comments above builtin definitions. Docs above definitions produce x.info.rye files (structures) that get evaluated by main.rye in tests (current) folder.

To find out more about the comment docs and format read this: https://ryelang.org/cookbook/improving-rye/one-source/

# Generating .info. files

To generate the info files out of builtins call `./regen`. The script uses /cmd/rbit/rbit tool (Go binary) that parses the Go code and creates the x.info.rye structures.

# Running tests

Once you use `regen` you can use the main.rye in this folder. To run it use the dot shortcut: `rye .`, this will show you help information.

```
# To list all tests groups
rye . ls

# Run all the table group tests
rye . test base
rye . test table
rye . test io

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
