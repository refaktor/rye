# Rye language 🌾

[![Build and Test](https://github.com/refaktor/rye/actions/workflows/build.yml/badge.svg)](https://github.com/refaktor/rye/actions/workflows/build.yml)
[![golangci-lint](https://github.com/refaktor/rye/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/refaktor/rye/actions/workflows/golangci-lint.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/refaktor/rye/badge)](https://securityscorecards.dev/viewer/?uri=github.com/refaktor/rye)
[![Go Reference](https://pkg.go.dev/badge/github.com/refaktor/rye.svg)](https://pkg.go.dev/github.com/refaktor/rye)
[![Go Report Card](https://goreportcard.com/badge/github.com/refaktor/rye)](https://goreportcard.com/report/github.com/refaktor/rye)
[![GitHub Release](https://img.shields.io/github/release/refaktor/rye.svg?style=flat)](https://github.com/refaktor/rye/releases/latest)
[![Homebrew](https://img.shields.io/homebrew/v/ryelang.svg?style=flat)](https://formulae.brew.sh/formula/ryelang)

**For comprehensive documentation, tutorials, and examples, visit [ryelang.org](https://ryelang.org/)**

## What is Rye?

Rye is a high-level, dynamic programming language inspired by Rebol, Factor, Linux shells, and Go. It features a Go-based interpreter and interactive console, making it an excellent scripting companion for Go programs. Rye can also be embedded into Go applications as a scripting or configuration language.

Key characteristics:
- **Homoiconic**: Code is data, data is code
- **Function-oriented**: No keywords, everything is a function call
- **Expression-based**: Everything returns a value
- **First-class functions**: Functions and code blocks are values
- **Multiple dialects**: Specialized interpreters for different tasks
- **Safety-focused**: Explicit state changes, pure/impure function separation, validation dialect

**Status**: Alpha - Core language design is stable, focus is on improving runtime, documentation, and usability.

## Quick Examples

```red
print "Hello World"

"Hello World" .replace "World" "Mars" |print
; prints: Hello Mars

"12 8 12 16 8 6" .load .unique .sum
; returns: 42

{ "Anne" "Joan" "Adam" } |filter { .first = "A" } |for { .print } 
; prints:
; Anne
; Adam

fac: fn { x } { either x = 1 { 1 } { x * fac x - 1 } }
; function that calculates factorial
range 1 10 |map { .fac } |print\csv
; prints: 1,2,6,24,120,720,5040,40320,362880,3628800

kind: "admin"
open sqlite://data.db |query { select * from user where kind = ?kind }
; returns: Table of admins

read %name.txt |fix { "Anonymous" } |post* https://example.com/postname 'text
; makes HTTP post of the name read from a file, or "Anonymous" if file failed to be read
```

For more examples and interactive demos, visit [ryelang.org/meet_rye/](https://ryelang.org/meet_rye/)

## Building Rye from Source

### Prerequisites

1. Install Go 1.21.5 or later:
   ```bash
   # Example for Linux
   wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
   rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin
   go version
   ```

### Clone and Build

1. Clone the repository:
   ```bash
   git clone https://github.com/refaktor/rye.git
   cd rye
   ```

2. Build options:

   - **Minimal build** (fewer modules, smaller binary):
     ```bash
     go build -tags "b_tiny" -o bin/rye
     ```

   - **Standard build** (most modules included):
     ```bash
     go build -o bin/rye
     # or simply
     ./build
     ```

   - **Custom build** (select specific modules):
     ```bash
     go build -tags "b_tiny,b_sqlite,b_http,b_json" -o bin/rye
     ```

3. Run Rye:
   ```bash
   # Run the REPL
   bin/rye
   
   # Run a script
   bin/rye script.rye
   ```

### Building WASM Version

Rye can run in browsers and other WASM environments:

```bash
GOOS=js GOARCH=wasm go build -tags "b_tiny" -o wasm/rye.wasm main_wasm.go
# or use the helper script
./buildwasm
# Then visit http://localhost:8085/ryeshell/
```

### Running Tests

```bash
cd info
../bin/rye . test

# Generate function reference
../bin/rye . doc
```

## Getting Rye (Pre-built)

If you prefer not to build from source, you have several options:

### Binaries

Pre-compiled binaries for **Linux**, **macOS**, **Windows**, and **WASM** are available under [Releases](https://github.com/refaktor/rye/releases).

### Package Managers

- **Homebrew** (macOS/Linux):
  ```bash
  brew install ryelang
  ```

- **ArchLinux User Repository** (Arch Linux):
  ```bash
  yay -S ryelang
  ```

### Docker Images

- **Binary image** (includes Rye and Emacs-nox):
  ```bash
  docker run -ti ghcr.io/refaktor/rye
  ```

- **Development image** (build from repository):
  ```bash
  docker build -t refaktor/rye -f .docker/Dockerfile .
  docker run -ti refaktor/rye
  ```

## Resources

- **[ryelang.org](https://ryelang.org/)** - Official documentation, tutorials, and examples
- **[Blog](https://ryelang.org/blog/)** - Latest updates and development news
- **[Examples folder](./examples/)** - Code examples and demos
- **[Asciinema demos](https://asciinema.org/~refaktor)** - Interactive terminal demos

## Extensions and Related Projects

- **[Rye-fyne](https://github.com/refaktor/rye-fyne)** - GUI toolkit binding
- **[Rye-gio](https://github.com/refaktor/rye-gio)** - Gioui toolkit binding (WIP)
- **[Rye-ebitengine](https://github.com/refaktor/rye-ebitengine)** - 2D game engine binding (WIP)
- **[ryegen](https://github.com/refaktor/ryegen)** - Binding generation toolkit (WIP)

## Editor Support

- **VS Code**: Search for "ryelang" in the Extension marketplace [repository](https://github.com/refaktor/rye-vscode)
- **Emacs**: [syntax highlighting](https://github.com/refaktor/rye/tree/main/editors/emacs)
- **NeoVIM**: [syntax highlighting](https://github.com/refaktor/rye/tree/main/editors/nvim)

## Community and Contact

- **[GitHub Discussions](https://github.com/refaktor/rye/discussions)**
- **[Reddit](https://reddit.com/r/ryelang/)**
- **[Issues](https://github.com/refaktor/rye/issues)**
- **Email**: janko.itm+rye[at]gmail.com
