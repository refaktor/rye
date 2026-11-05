# Rye Syntax Checklist (v20260917)

This is a condensed guide for writing correct Rye code, based on:
- Current Rye (as of repo state) behaviour
- Existing Rosetta Code examples (all verified running)

---

## 0. TL;DR rules you must not forget
- Inside loops/maps, never use setwords (`name:`) or `var` repeatedly. Use modword assignment right (`name:: value`) or left (`value ::name`).
- Op-words (+, -, *, /, =, >, <, <mod>, <concat>, etc.) are right-associative and same priority.
- `+` is numeric only. For strings use `++` (two strings) or `concat` or `join { "block" "of" "strings" }`.
- `if` has one block only. Use `either` for if/else.
- `integer` parses number strings only. No char→int (`char` and `ascii` do that).
- `decimal` takes strings; for int→decimal, multiply by `1.0`. #TODO - this will be fixed
- Commas (expression guards) can separate expressions at block level; you can’t place a comma inside an expression.
- Setwords (`x:`) are immutable forever; `::` cannot modify them later.
- `failure "msg"` creates an **Error value**; `fail "msg"` / `ensure` set **failed state**. Only `fail`-style failures abort surrounding `map`/`filter`/`fold`; wrap such calls in `try` or `|fix` when iteration should continue.
- `block list-value` converts a List to a Block (there is no `to-block` builtin). `unpack` flattens one level; `split\every` chunks.
- Prefer non-builtin parameter names; avoid shadowing builtins like `lower`, `split`, `print`, etc.
- A setword can be defined only once. During a rewrite, keep one active name and use a distinct reference name like `f-OLD`.
- Don't put `^check`/`^ensure` with a concatenated message inside a setword capture; capture first, then check on its own line.

---

## 1. Bindings: setword vs var vs modword
- Setword `x:` creates an immutable binding. It cannot be changed by `::` later.
- Var creates a mutable name: `var 'x 0` then `x:: 1`.
- Modword creates or modifies a mutable binding: `y:: 42` then `y:: 43`.
- Setwords and modwords also have left variants that take the value from the left: `12 :z`, `44 ::y`.
- In loops/maps, assign with a modword: `... { ::n , n * 2 ::sum }` or `n * 2 ::sum`.
- `::` affects the current context. Inside a `fn`, it creates/updates a local binding; to mutate an existing top-level `var`, use `change! new-value 'name`.
- A name can be set only once; a second `name:` at the same level is a runtime error.

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
- `.f arg` injects the left value as the first argument; `.f* arg` injects it as the last argument. Use `.f*` when iterating candidates for a function like `fn { s fmt } { ... }`.

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
- Common ops: `length?`, `index?`, `substring`, `.upper`, `.lower`, `.reverse`, `.contains`, `replace`, `.split`, `has-prefix`, `has-suffix`.
- Prefix/suffix checks: `has-prefix s "-"`, `s .has-prefix "-"`, `has-suffix f ".txt"`.
- Character iteration: `"text" .map { ::ch , ... } |join`.
- No built-in char→int; for Caesar/ROT-13 use two alphabets and `index?`.
- Regex split: use `split s regex "\s+"`, not `s .split regex "\s+"` (the method form does not dispatch the regex argument correctly).

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
- Lists vs blocks: `block list-value` converts a List to a Block. (There is no `to-block` builtin.) Use `ref` for a mutable copy.
- `unpack` flattens a block of blocks one level. `split\every coll n` chunks a block or string into blocks/strings of size `n`.
- `map\idx coll 'i { ... }` and `for\idx coll 'i { ... }` expose a 0-based index. The `\pos` variants (`map\pos`, `for\pos`) use 1-based positions.
- `head coll n` / `tail coll n` keep the first/last `n`; negative `n` drops from the other end. `rest\from coll n` drops the first `n` and fails if `n` is past the end.
- `seek coll { pred }` and `any coll { pred }` return the matching **element**, not the predicate result. Use `try`/`|fix { false }` when a miss is normal.
- `sort\by` uses a boolean comparator. It is **not stable**: add explicit tie-breakers when input order matters.
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
- **Error value, no failed state:** `failure "msg"`, `failure\wrap "outer" err`.
- **Error value + failed state:** `fail "msg"` (sets the failure flag), `ensure "msg"` (sets it when the condition is false).
- **Error value + failed state + early return:** `^fail "msg"`, `^check "msg"`, `^ensure "msg"`.
- `is-error v` tests for an Error value. `try { ... }` converts failed state into an Error value, so it can be inspected.
- Handle with `|fix { fallback }`, `|^fix { fallback }`, `|check "ctx"`, `|^check`, `|ensure "msg"`, `|^ensure`.
- Control flow on success: `|continue { ... }`; disarm: `try { ... }` or `|disarm`.
- Extras: `fix\either`, `fix\else`, `retry { }`, `finally { op } { cleanup }`.
- Use some care with `^`-variants: they set the return flag. In a `map`/`filter`/`fold`, they can cancel the whole iteration. If you want an error **value** per item, use `failure` or wrap the call in `try`.
- Do not put a `^`-variant after an expression whose result is also being captured by a setword, especially with `++`; capture and check on separate statements:
  ```
  t: parse-date s fmt
  t |^check ( "does not match format " ++ fmt )
  ```
- `|fix` handles failed state, but not the early-return flag from `^check`/`^ensure`. Use `try { helper ... }` as the boundary when iterating.
- Typical split: fail fast with `^check`/`^ensure` in leaf parsers, then catch with `try` where the caller expects `is-error`.

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
- Pipeline validation and parsing:
  ```rye
  parse-delta: fn { s } {
      .trim :t .length?
      |_>= 2 |^ensure "Not a valid delta: " ++ s

      unit: tail t 1
      num: head t -1 |integer |^check "Not a valid delta: " ++ s

      num |> 0 |^ensure "Not a valid delta (must be positive): " ++ s
      num * switch unit { "s" { 1 } "h" { 3600 } "d" { 86400 } _ { 0 } }
  }
  ```
- When a fail-fast parser is used inside `map`, catch it per item if you want to collect errors:
  ```rye
  ds: map items { ::it , try { parse-delta it } }
  bad: filter ds { ::d , is-error d }
  ```

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
- Any loop-local names assigned with a modword, not `name:` or `var`?
- Any chained op-words parenthesized where needed?
- Any commas placed only between statements, not inside expressions?
- Parameters not shadowing builtins?
- Integer vs decimal division intentional?
- Using `ref { }` for collections you mutate?
- `failure` vs `fail`/`^ensure` chosen intentionally?
- `fail`/`ensure` inside `map`/`filter`/`fold` wrapped in `try` if iteration should continue?
- `block` used for List→Block, not `to-block`?
- `head`/`tail` not mistaken for “drop first n”; use `rest\from` for that?
- `sort\by` stability assumptions made explicit with tie-breakers?
- Only one active definition per setword (no duplicate `f:` and `f:` while keeping `f-OLD` for reference)?
- `^check`/`^ensure` messages with `++` parenthesized, and checks on their own statement rather than inside a capture?
- `any`/`seek` result treated as an element, not as a boolean predicate result?
- Global mutable updated with `change!`, local mutable updated with `::`?
- Regex split written as `split s regex r`?

If all yes, your Rye should run.

---

Appendix A: Empty values quick ref
- `first { }`, `last { }`, `rest { }`, `reduce { }` fail.
- `fold { } init fn { item acc } { ... }` returns `init`.
- `fold\do { } 'acc init { ... }` returns `init`.
- `sum { }` ⇒ 0; `map`/`filter` on empty return empty; `length? { }` ⇒ 0.

Appendix B: Pipe equality family
- `|=`, `|>`, `|<`, `|>=`, `|<=`, `|!=` ensure each side is evaluated then compared.
- Pipe comparisons also have `_` variants such as `|_>=`, used with injected values: `length? xs |_>= 2`.

Appendix C: Error value vs failed state
- `failure "x"` → Error value, no flags.
- `fail "x"` → Error value + failure flag.
- `^fail "x"` → Error value + failure + return flags.
- `ensure` / `ensure\with` → failure flag on false.
- `^ensure` / `^check` → failure + return flags on failure.
- `try { expr }` → returns the result, or an Error value if failed state was raised.

Appendix D: Date/time builtins
- `parse-date s fmt` uses strptime formats: `parse-date "15/01/2023" "%d/%m/%Y"`.
- `format-date t fmt` uses strftime formats: `now |format-date "%Y-%m-%d %H:%M:%S"`.
- `format-date-go t layout` preserves Go layout syntax when needed.
- A strptime parse without a timezone produces a UTC time; `now` is local.
- `parse-date` currently requires cgo and may accept partial input. Use a round-trip check when exact matching matters:
  ```rye
  back: format-date t fmt
  back |= s |^ensure ( "does not match format " ++ fmt )
  ```

End of guide.
