# Generated reference

This is next, w-i-p folder for tests and builtin reference. Both are generated from evaldo/builtins*.go files.

For more about this read: https://ryelang.org/cookbook/improving-rye/one-source/

## Use

to generate do in cmd/parse_bultins folder:

    ./parse_builtins ../../evaldo/builtins.go > ../../tests2/base.tests.rye
    ./parse_builtins ../../evaldo/builtins_table.go > ../../tests2/table.tests.rye


To run tests do in this folder:

    rye . test base
    rye . test table

The generator, testing framework and tests and additional info is being improved right now. This is a very temporary state.

