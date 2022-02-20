# Rye TinyGo WASM test

This is a codebase of go, ad-hoc cleaned of all dependencies, even the parser/loader that depends on go-peg library for first test.

## To build

```bash
# There are no dependencies so no go gets
# Project doesn't use go-modules yet, since modules you need
# are absolutely dependant on the build flags you provide and
# I haven't seen yet how to specify this with modules
export GO111MODULE=auto
# Build with regular Go
GOARCH=wasm GOOS=js go build
# Or with TinyGo
tinygo build -target wasm -size full -o main
```
