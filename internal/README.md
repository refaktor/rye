# Internal Tests

These are internal tests for developing the runtime and the loader (parser).

Tests of builtin functions can be found in `info/` folder. `cd ../info`. You can then run them by `rye . test`.

## Test Organization

### Subfolder Tests (assert\display pattern)

Most test subfolders use the `assert\display` pattern with `private load` for context isolation:

```bash
cd internal/basics && rye main.rye
cd internal/parsing && rye main.rye
cd internal/evaluator && rye main.rye
cd internal/closures && rye main.rye
cd internal/failure_handling && rye main.rye
cd internal/cli-parse && rye main.rye
cd internal/structures && rye main.rye
cd internal/observer && rye main.rye
```

### Top-Level Framework (internal/main.rye)

The top-level `internal/main.rye` uses a custom test framework with `section`, `group`, `equal`, `error`, `stdout` functions. This is used by the `*.rye` files directly in this folder:

- `basics.rye`
- `misc.rye`
- `structures.rye`
- `validation.rye`

Run with: `rye . test` or `rye . test basics`

**Note:** This is a separate system from the subfolder tests and is kept for historical/documentation generation purposes.

### Special Folders

- `patterns/` - Has its own framework for documentation generation
- `unshare/` - Requires `--unshare` flag, see `unshare/README.md`
- `persistent/` - Partially implemented, see `persistent/README.md`
- `from_website/` - Future: auto-extracted website examples
- `error_reporting/` - Test file generator for error reporting
- `tpl/` - Templates for documentation generation
- `go_tests/` - Go test files (not Rye tests)

## Running All Subfolder Tests

```bash
# Run each subfolder's main.rye
for dir in basics parsing evaluator closures failure_handling cli-parse structures observer; do
  echo "=== $dir ==="
  cd internal/$dir && rye main.rye && cd ../..
done
```
