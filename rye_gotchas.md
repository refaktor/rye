# Rye Gotchas and Confusing Aspects

This guide highlights language aspects that commonly trip up users, with minimal examples of the wrong and right patterns.

## 1) Setwords vs var vs modword (especially in loops)

- Setwords (`name:`) are immutable; `var` creates mutable; modword (`::`) modifies/creates mutables.
- Inside loops, do not use setwords or re-declare with `var` on each iteration.

Wrong:
```
for range 1 3 { ::i
  result: i * 2
}
```

Wrong:
```
for range 1 3 { ::i
  var 'result i * 2
}
```

Right:
```
for range 1 3 { ::i
  i * 2 ::result
}
```

Also wrong:
```
x: 5
x:: 10  ; cannot modify setword
```

Right:
```
var 'x 5
x:: 10
```

## 2) Operator precedence and associativity (op-words vs pipe-words)

- Op-words (+, -, *, /, =, .mod, etc.) have equal priority and are right-associative.
- Pipe-words (`|word`) run left-to-right and wait for the left expression to finish.

Wrong:
```
if n .mod d = 0 { ... }    ; parsed as n .mod (d = 0)
```

Right:
```
if ( n .mod d ) = 0 { ... }
```

Right (pipe for ordering):
```
1 + 2 |* 3                  ; (1 + 2) * 3 = 9
```

Another trap:
```
d * d > n                   ; parsed as d * (d > n)
```

Right:
```
sq: d * d , sq > n
```

## 3) Comparing results of function calls: use pipe equality

- Bare `=` is an op-word and eagerly grabs the next token.
- When both sides are expressions, especially starting with a word, use `|=`.

Wrong:
```
sorted lhs = sorted rhs
```

Right:
```
sorted lhs |= sorted rhs
```

## 4) Commas are not list separators; they separate statements

- A comma cannot appear inside a single expression tree. Use temporaries or split statements.

Wrong:
```
join { compute-a , compute-b }
```

Right:
```
a: compute-a , b: compute-b , join { a b }
```

Re-injection example:
```
value .with {
  * 10 |print ,    ; re-injects original value into next statement
  + 10 |print
}
```

## 5) Strings: + is numeric only; conversions are explicit

- `+` does not concatenate strings; use `++` or `join`.
- `join` requires all items to be strings.

Wrong:
```
"Hello " + "World"
```

Right:
```
"Hello " ++ "World"
join { "Hello" " " "World" }
```

Wrong:
```
print join { "Count: " 42 }  ; non-string in join
```

Right:
```
print join { "Count: " string 42 }
```

## 6) No built-in char→int; use index? lookups

- There’s `to-char` (int→char), but no char→int. Using `to-integer` on characters fails.

Wrong:
```
to-integer "A"
```

Right (ROT13-like approach):
```
alpha: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
coded: "NOPQRSTUVWXYZABCDEFGHIJKLMnopqrstuvwxyzabcdefghijklm"
"Hello" .map { ::ch
  index? alpha ch ::i
  either i >= 0 { substring coded i i + 1 } { ch }
} |join
```

## 7) if vs either (no else in if)

- `if` only takes one block; the next block is not an else.

Wrong:
```
if n > 0 { print "pos" } { print "neg" }
```

Right:
```
either n > 0 { print "pos" } { print "neg" }
```

## 8) Injection model in blocks (for, map, with, loop)

- In injected blocks, pick up the injected value with an op-word or capture it with `::name`.

Examples:
```
with "hello" { .upper .print }
for { 1 2 3 } { + 10 |print }          ; each element injected
"ab" .map { ::ch , ch } |join
```

Common pitfall:
```
for range 1 3 { print + 10 }           ; print tries to collect + and 10 as args
```

## 9) Lists vs blocks; mutation requires ref

- Blocks `{ }` are immutable; to mutate, use `ref { }`.

Wrong:
```
items: { 1 2 3 }
append! 4 'items
```

Right:
```
items: ref { 1 2 3 }
append! 4 'items
update! items 0 99
items -> 0  ; 0-based index
```

Wrong:
```
to-block list { 1 2 3 }  ; no such builtin
```

Right:
```
l: list { 1 2 3 }
block l                  ; List -> Block
```

## 10) Function parameters that shadow builtins

- Using parameter names that are also builtins can cause evaluation issues.

Wrong:
```
normalize: fn { text } { lower text }
```

Right:
```
normalize: fn { txt } { lower txt }
```

Avoid: `lower`, `upper`, `split`, `join`, `print`, `map`, `filter`, `reduce`, `length?`, `sum`, `min`, `max`, `sort`, `reverse`, `text`, `value`, `list`, `block`, `type`, etc.

## 11) `failure` vs `fail` vs `ensure` (error value vs failed state)

This is one of the easiest things to mix up.

- `failure "msg"` creates an **Error value** and sets no flags.
- `fail "msg"` creates an Error value and sets the **failure flag**.
- `fail` is roughly `false |ensure "msg"`.
- `ensure` sets the failure flag when its condition is false.
- `^fail`, `^check`, and `^ensure` also set the **return flag**, so they unwind the current function immediately.
- `try { ... }` turns failed state into an inspectable Error value.
- `is-error v` only checks whether `v` is an Error value.

```rye
; value: callers can inspect this with is-error
value-error: fn { x } {
  either x > 0 { x * 10 } { failure "bad" }
}

; failed state: aborts the surrounding map/filter/fold
fail-fast: fn { x } {
  x > 0 |^ensure "bad"
  x * 10
}
```

This is important inside higher-order functions:

```rye
map { 2 -2 3 } { ::x , value-error x }
; [Block: ^[Integer: 20] [Error: bad] [Integer: 30]]

map { 2 -2 3 } { ::x , fail-fast x }
; ⚠ FAILURE: Error: bad
```

If you want iteration to continue, catch the failure per item:

```rye
map { 2 -2 3 } { ::x , try { fail-fast x } }
; [Block: ^[Integer: 20] [Error: bad] [Integer: 30]]

map { 2 -2 3 } { ::x , fail-fast x |fix { failure "wrapped" } }
; [Block: ^[Integer: 20] [Error: wrapped] [Integer: 30]]
```

Rule of thumb:

- Use `failure` when the Error is **data** that callers inspect.
- Use `fail`/`ensure`/`^ensure` when you want to **abort the current operation**.
- If a fail-fast function is called inside `map`/`filter`/`fold` and you want per-element errors, wrap it in `try` or `|fix`.

Example of fail-fast validation:

```rye
fetch: fn { id } {
  id > 0 |^ensure "ID must be positive"
  get uri |^check "HTTP failed"
  |json/parse |^check "Bad JSON"
}
```

Breaking a forever:

```rye
forever {
  ^if ready? { "done" }
}
```

## 12) Integer vs decimal math (division)

- `10 / 3` = 3 (integer). Promote to decimal explicitly.

Wrong:
```
10 / 3
```

Right:
```
10 / 3.0
10 * 1.0 / 3
```

## 13) Map/filter/reduce on empty collections

- Some operations return identity/empty; others fail.

Examples:
```
sum { }                             ; 0
map { } { + 1 }                     ; { }
filter { } { > 0 }                  ; { }
first { }                           ; failure
reduce { } 'a { + a }               ; failure
fold { } 0 fn { item acc } { item + acc }   ; 0
fold\do { } 'acc 0 { + acc }        ; 0
```

Note the two `fold` forms:

- `fold coll initial fn { item acc } { ... }`
- `fold\do coll 'acc initial { ... }`

## 14) Mixed chaining of .methods and |pipes

- `.methods` are op-words. When mixing with pipes, precedence can surprise.

Examples:
```
"Hello World" .lower .split " " |for { .print }
1 + 2 |* 3    ; pipes bind after +, so it’s (1 + 2) * 3
```

Safer:
```
part: "Hello World" .lower .split " "
for part { .print }
```

## 15) Left variants of set/modwords

- Left assignments exist and capture the left value; they can be surprising.

Examples:
```
12 :z     ; z is setword with 12
44 ::y    ; creates/modifies mutable y with 44
```

Tip: Prefer right variants unless left variants clearly improve readability.

## 16) Joining many values vs two

- For two strings: `++`.
- For many parts or mixed types: `join` with explicit string conversions.

Examples:
```
name: "Jim" , age: 42
print join { "Name: " name " Age: " string age }
"Hi, " ++ name
```

## 17) Shadowing vs using injected names incorrectly

- In injected blocks, capture the value with `::name` if you’ll reuse it.

Less ergonomic:
```
"abc" .map { index? "abc" .print , index? "abc" }
```

Better:
```
"abc" .map { ::ch , index? "abc" ch ::i , i }
```


## 18) `head` / `tail` / `rest\from`: keep vs drop

`head` and `tail` mean “keep the first/last n”, not “drop the first/last n”.
Negative counts drop from the other end.

```
head { 1 2 3 4 } 2       ; { 1 2 }
tail { 1 2 3 4 } 2       ; { 3 4 }

head "abcd" -1           ; "abc"  (all but the last character)
tail "abcd" -1           ; "bcd"  (all but the first character)
```

To drop the first n elements, use `rest\from`:

```
rest\from { 1 2 3 4 } 2 ; { 3 4 }
```

But `rest\from` fails when asked to start past the end:

```
try { rest\from { 1 2 } 5 } |type?  ; 'error
```

If the skip may reach the end, guard it:

```
drop: fn { xs n } { either n >= length? xs { { } } { rest\from xs n } }
```

## 19) `sort\by` is not stable

`sort\by` uses a boolean comparator, but it is not a stable sort. Do not assume
equal elements stay in input order.

```
sort\by { 6 12 1 } fn { a b } { a < b }   ; { 1 6 12 }
sort\by { 6 12 1 } fn { a b } { a > b }   ; { 12 6 1 }
```

If input order matters for ties, carry an index and make it an explicit tie-breaker:

```
with-index: fn { items } { map\idx items 'i { ::it , [ it i ] } }

entries: with-index backups
sorted: sort\by entries fn { a b } {
    either ( second first a ) = ( second first b ) {
        ( second a ) < ( second b )
    } {
        ( second first a ) > ( second first b )
    }
}
```

This matters when porting Python code that relies on `sorted(..., reverse=True)`
or `sorted(..., key=...)` being stable.

## 20) Date/time: strptime, strftime, UTC, and cgo

Rye now has native strptime/strftime-style builtins:

```
parse-date "15/01/2023" "%d/%m/%Y"
now |format-date "%Y-%m-%d %H:%M:%S"
now |format-date-go "2006-01-02"   ; old Go layout syntax
```

Two practical traps:

- A strptime parse without a timezone yields a UTC time, while `now` is local.
- `parse-date` currently requires cgo and the BSD strptime implementation may accept
  partial/trailing input. If exact matching matters, round-trip check it:

```
t: parse-date s fmt
back: format-date t fmt
back |= s |^ensure "does not match format " ++ fmt
```

For UTC “now”, parse a formatted local wall-clock string:

```
utc-now: fn { } {
    iso: format-date now "%Y-%m-%dT%H:%M:%S"
    parse-date iso "%Y-%m-%dT%H:%M:%S"
}
```

## 21) CLI dialect: repeatable options are “one value per occurrence”

The CLI dialect is useful, but it is not `argparse`:

```
spec: {
  program "tool"
  -v|verbose flag
  -s|sources string list
  command: string required
  jobs: string many
}

res: parse-args { -s a -s b make one two } spec
```

- `list` collects **one value per flag occurrence**. `-s a b` does not put both
  `a` and `b` into `sources`; the second value becomes positional.
- `many` collects all remaining positional values.
- There is no built-in `-` terminator for “stop this list”. If you support
  argparse-style `--sources a b - make`, normalize the args first or require the
  repeated form.
- Parsed `list`/`many` values are List values, not Blocks. Use `block` before
  calling block-oriented builtins such as `contains`:
  ```
  sources: block res."sources"
  ```
- `parse-args\ctx` returns a Context instead of a Dict, which can be nicer for
  read-only access.
- `generate-help spec` and `format-parse-errors err` make the dialect practical
  for real tools.
- As with other fail-fast helpers, `try { parse-args ... }` gives you an Error
  value instead of a failed program state.

## 22) A setword can only be defined once

Setwords are constants. Re-evaluating `name:` at the same level is a runtime error:

Wrong:
```
f: fn { } { 1 }
f: fn { } { 2 }      ; Error: f is already set and constant
```

During a rewrite, use a distinct reference name (`old-f`, `f-OLD`) and keep only one active definition. Do not leave two definitions with the same name.

## 23) `^check` / `^ensure` inside a captured expression

`^`-variants set the return flag. If one appears on the right side of a setword capture and the message is concatenated, Rye can capture the message string instead of an Error object, and the concatenation may happen outside the check.

Wrong:
```
t: parse-date s fmt |^check "bad date " ++ fmt
```

Right:
```
t: parse-date s fmt
t |^check ( "bad date " ++ fmt )
```

Same for a round-trip condition:

```
back: format-date t fmt
back |^check ( "does not match " ++ fmt )
back |= s |^ensure ( "does not match " ++ fmt )
```

For a helper that uses `^check`, wrap the call in `try` when the caller wants an Error value or wants to keep iterating:

```
parsed: map formats { ::f , try { parse-date-fmt s f } }
found: filter parsed { ::r , not is-error r }
```

Do not rely on `|fix` to catch the early-return flag; `try` is the explicit boundary.

## 24) `seek` / `any` return the matching element, not the predicate result

```
seek { "no" "ok" } { = "ok" }      ; "ok"
any { 1 2 3 } { > 1 }              ; 2
```

If the predicate computes something (a parsed value, a boolean), call it again, keep pairs, or map first:

```
attempts: map formats { ::f , [ f try { parse-date-fmt s f } ] }
match: seek attempts { ::a , not is-error second a }
either is-bool match { false } { second match }
```

When a miss is normal, guard it: `try { seek ... }` or `seek ... |fix { false }`.

## 25) `.fn*` appends the injected value; `.fn` prepends it

```
f: fn { a b } { a ++ ":" ++ b }
"L" .f "R"       ; "L:R"
"L" .f* "R"      ; "R:L"
```

This is handy when iterating candidates and the injected value belongs in the last argument:

```
; calls parse-date-fmt s f for each candidate f
seek formats { ::f , try { .parse-date-fmt* s } }
```

## 26) Modwords in a function do not update a top-level var

`x:: value` creates or modifies a binding in the current context. Inside a `fn`, that is a local binding; it does not reach out and update a top-level `var`.

```
var 'count 0
bump: fn { } { count:: count + 1 }
bump
count              ; still 0
```

Use `change!` for an existing outer/global mutable:

```
bump: fn { } { change! count + 1 'count }
```

Use `::` for locals and for variables declared in the same context.

## 27) Regex splitting uses the function form

`split` accepts a compiled regex as its second argument, but the method form does not inject the string into the right position:

Wrong:
```
"a  b" .split regex "\s+"
```

Right:
```
split "a  b" regex "\s+"
; or
words-of: fn { s } { split s regex "\s+" }
```

The same method-injection rule applies to other builtins with multiple argument kinds: when dispatch fails, try the plain function form with explicit arguments.
