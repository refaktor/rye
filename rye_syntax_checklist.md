# Rye Syntax Checklist (v20260917)

This is a condensed guide for writing correct Rye code, based on:
- Current Rye (as of repo state) behaviour
- Existing Rosetta Code examples (all verified running)

---

## 0. TL;DR rules you must not forget
- Inside loops/maps, never use setwords (`name:`) or `var` repeatedly. Use modword assignment: `value ::name`.
- Op-words (+, -, *, /, =, >, <, <mod>, <concat>, etc.) are right-associative and same priority.
- `+` is numeric only. For strings use `++` (two strings) or `concat` or `join { "block" "of" "strings" }`.
- `if` has one block only. Use `either` for if/else.
- `integer` parses number strings only. No char→int (`char` and `ascii` do that).
- `decimal` takes strings; for int→decimal, multiply by `1.0`. #TODO - this will be fixed
- Commas (expression guards) can separate expressions at block level; you can’t place a comma inside an expression.
- Setwords (`x:`) are immutable forever; `::` cannot modify them later.
- Prefer non-builtin parameter names; avoid shadowing builtins like `lower`, `split`, `print`, etc.

---

## 1. Bindings: setword vs var vs modword
- Setword `x:` creates an immutable binding. It cannot be changed by `::` later.
- Var creates a mutable name: `var 'x 0` then `x:: 1`.
- Modword creates or modifies a mutable binding: `y:: 42` then `y:: 43`.
- Set and Modwors come with their left variants that tale vale from the left `12 :z` , `44 ::y`
- In loops/maps, always assign with modword: `... { ::n , n * 2 ::sum }`.

---

## 2. Evaluation and operators
Priority (high→low):
1) setword/modword capture; 2) pipe words (`|word`) left-to-right; 3) op-words (`+`, `*`, `=`, `.op`) right-associative; 4) regular words collect args from right.

Implications:
- Chain of op-words parses rightward: `a + b * c` ⇒ `a + (b * c)`; `a .mod b = 0` ⇒ `a .mod (b = 0)` which is wrong → use `(a .mod b) = 0`.
- Prefer pipe equality when comparing two computed expressions: `left |= right` (also for `|>`, `|<`, etc.).
- Set/modwords grab the entire right side: `x:: 1 + 2 * 3` assigns 7.

Pipes and methods:
- `5 |+ 3`, `"abc" |join\with ","`, `"Hello" .upper`.
- Chain safely with pipes when mixing operations: `1 + 2 |* 3` ⇒ `(1 + 2) * 3`.

---

## 3. Blocks, commas, and injection
- Commas separate statements within a block; not allowed inside a single expression node. Use temporaries:
  - Wrong: `join { a , b }` inside a single expression
  - Right: `a: expr1 , b: expr2 , join { a b }`
- Many forms inject a value into a block: `with v { .print }`, `for col { ::x ... }`, `map`, `filter`, `reduce`, `loop`.
- A comma inside an injected block re-injects the original value for the next statement.
- `fn1 { body }` defines a single-arg function that uses injection.

---

## 4. Control flow
- `if cond { block }` only has a true branch.
- `either cond { T } { F }` for if/else.
- `cases val { {pred} {..} _ {default} }` and `switch`/`match` are available.
- `^if cond { value }` returns from the enclosing function early.
- `while { cond } { ... }`, `loop n { ... }` (injects 1..n), `for col { ::x ... }`, `forever { ... }` (use `^if` or returns to exit).

Indexed loops:
- `for\idx coll 'i { ::v ... }` has index word and injected value.

---

## 5. Strings
- Use `++` to append two strings, or `join { ... }` for many.
- Common ops: `length?`, `index?`, `substring`, `.upper`, `.lower`, `.reverse`, `.contains`, `replace`, `.split`.
- Character iteration: `"text" .map { ::ch , ... } |join`.
- No built-in char→int; for Caesar/ROT-13 use two alphabets and `index?`.

---

## 6. Numbers and math
- Integer division truncates: `10 / 3` ⇒ 3. Use a decimal to get fractional: `10 / 3.0`.
- `%` (or `.mod`) for modulo.
- Decimal math functions live in `math` context: `do\in math { sqrt 9 }`, `math/pi`.
- Conversions: `string` (was `to-string`), `to-integer "42"`, `to-decimal "3.14"`, `n * 1.0` for int→decimal, `to-char 65`.

---

## 7. Collections and mutation
- Immutable block: `{ 1 2 3 }`. Mutable ref-block: `ref { 1 2 3 }` then `append! 4 'blk`, `update! blk 0 99`, `blk -> 0`.
- `replicate n { value }` creates repeated values.
- Lists vs blocks: `to-block` converts a List to Block; don’t call it on a Block. Use `ref` for a mutable copy.
- Empty collection behaviour: `first { }`/`last { }` fail; `sum { }` ⇒ 0; `map`/`filter` over empty return empty.

---

## 8. Functions and closures
- `fn { a b } { body }`, `does { body }`, `fn1 { body }`, `closure { a } { body }`, `fn\cc` captures current context.
- Returns last expression value in body.
- Closures capturing outer mutable vars: `var 't 0 , adder: closure { x } { t:: t + x , t }`.
- Avoid `fn\inside` with `context { }` (can create circular refs). Avoid `change!` on captured vars; use `inc!`/modword.
- Avoid parameter names that shadow builtins: e.g., `lower`, `split`, `print`, `map`, `filter`, `reduce`, `length?`, `min`, `max`, `sum`, `sort`, `reverse`, and also general names like `text`, `value`, `list`, `block`, `type`.

---

## 9. Contexts
- Create: `ctx: context { name: "Jim" , greet: does { print join { "Hi, I'm " name } } }`.
- Access: `ctx/name`, `do\in ctx { greet }`.
- Inheritance: `extends`; isolation: `isolate` with whitelisted words.
- Parent access: `@/name`, grandparent: `@/@/name`.
- `bind!` can rebind parent of a context.

---

## 10. Generic methods (Kinds) and URIs
- Kinds derive from value form, e.g., `sqlite://main.db` (sqlite-uri), `%file.txt` (file-uri), `https://...` (https-uri).
- Generic methods are capitalized: `Open`, `Read`, `Write`, `Get` etc., dispatch on first arg kind.
- Discover methods in console with `lg`.

---

## 11. Failures and error handling
- Failures are values. Create: `fail "msg"`, `^fail "msg"`, `failure "msg"`, `refail`, `failure\wrap`.
- Handle: `|fix { fallback }`, `|^fix { fallback }`, `|check "ctx"` (or number), `|^check`, `|ensure "msg"`, `|^ensure`.
- Control flow on success: `|continue { ... }`; disarm: `try { ... }` or `|disarm`.
- Extras available: `fix\either`, `fix\else`, `retry { }`, `finally { op } { cleanup }`.
- Returning words with `^` both act and return from current function.

---

## 12. Rosetta Code patterns you can reuse safely
- Factorial (recursive):
  `factorial: fn { n } { either n < 2 { 1 } { n * factorial n - 1 } }`
- Summation loop:
  `var 's 0 , for range 1 100 { ::n , s:: s + n } , s`
- Per-character transform with lookup:
  `alpha: "ABC...abc..." , coded: "NOP...nop..." , text .map { ::ch , index? alpha ch ::i , either i >= 0 { substring coded i i + 1 } { ch } } |join`
- Build list by condition: `res: ref { } , for range 1 100 { ::n , if .is-even { append! n 'res } }`
- Sieve: `flags: ref replicate size { true } , update! flags i false , flags -> i`

---

## 13. Testing pattern
- Use `assert\display "desc" { expr } expected`.
- Error tests: `assert\display "fails" { try { bad } |type? } 'error`.
- Grouping with printed headers. Load helpers via `private Load %file.rye` in a main test file.

---

## 14. Style
- kebab-case identifiers; comments start with `;`.
- Use tabs for indentation.
- Rosetta files should start with `; # Rosetta Code: Task Name` and end with expected output as a comment.

---

## 15. Gotchas checklist before you run
- Any `=` between two computed expressions replaced with `|=`?
- Any arithmetic/string mixing handled via `string` and `join`/`++`?
- Any loop-local names assigned with `::name`, not `name:` or `var`?
- Any chained op-words parenthesized where needed?
- Any commas placed only between statements, not inside expressions?
- Parameters not shadowing builtins?
- Integer vs decimal division intentional?
- Using `ref { }` for collections you mutate?

If all yes, your Rye should run.

---

Appendix A: Empty values quick ref
- `first { }`, `last { }`, `rest { }`, `reduce { }` fail.
- `fold { } 'acc init { ... }` returns `init`.
- `sum { }` ⇒ 0; `map`/`filter` on empty return empty; `length? { }` ⇒ 0.

Appendix B: Pipe equality family
- `|=`, `|>`, `|<`, `|>=`, `|<=`, `|!=` ensure each side is evaluated then compared.

End of guide.
