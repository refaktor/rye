# Rye Language - VS Code Extension

Syntax highlighting for the [Rye programming language](https://ryelang.org).

## Features

Comprehensive syntax highlighting for all Rye language constructs:

- **Comments** — `;` line comments
- **Strings** — `"double-quoted"` with escape sequences and `` `backtick raw strings` ``
- **Numbers** — integers and decimals
- **Constants** — `true`, `false`, `none`, `nil`, `_` (void)
- **URIs** — `https://example.com`, `file://path`
- **File paths** — `%file.txt`, `%.config`
- **Email addresses** — `user@domain.com`
- **Set-words** — `word:` (assignment), with special highlighting for function definitions
- **Mod-words** — `word::` (reassignment)
- **Left set/mod words** — `:word`, `::word` (injection)
- **Get-words** — `?word` (value reference without calling)
- **Tag-words** — `'word` (quoted symbols)
- **Op-words** — `.word` (method-style piped calls)
- **Pipe-words** — `|word` (pipe function calls)
- **Return words** — `^word` (early return patterns like `^fail`, `^check`)
- **XWords** — `<integer>`, `<string>` (type tags)
- **Context paths** — `module/function`, `term/blue`
- **Flag words** — `--verbose`, `-v`
- **Operators** — `+`, `-`, `*`, `/`, `//`, `++`, `->`, `<-`, `=`, `>`, `<`, `>=`, `<=`, `and`, `or`, `not`, `xor`
- **Keywords** — function definition (`fn`, `does`, `closure`), control flow (`if`, `either`, `switch`, `for`, `loop`, `while`), error handling (`try`, `fail`, `fix`, `check`), contexts (`context`, `do\in`, `private`), and more
- **Builtins** — 200+ built-in functions organized by category
- **Bracket colorization** — `{}`, `[]`, `()`
- **Code folding** — block-based folding
- **Auto-closing pairs** — for all bracket types and strings

## Installation

### From source (development)

1. Copy or symlink this directory to your VS Code extensions folder:
   ```bash
   ln -s /path/to/rye/editors/vscode ~/.vscode/extensions/ryelang
   ```
2. Restart VS Code or run "Developer: Reload Window"

### From VSIX

```bash
cd editors/vscode
npx @vscode/vsce package
code --install-extension ryelang-0.1.0.vsix
```

## Marketplace

The VSCode plugin also lives in its own repository: https://github.com/refaktor/rye-vscode

Install it by searching **Rye** or **Ryelang** in the VSCode marketplace.

## Supported File Extensions

- `.rye`
