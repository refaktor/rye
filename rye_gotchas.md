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
to-block { 1 2 3 }  ; already a Block
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

## 11) Early return and “returning words” (^)

- `^if` and `^check`/`^ensure`/`^fix` perform their action and return from the enclosing function.

Example:
```
fetch: fn { id } {
  id > 0 |^ensure "ID must be positive"
  get uri |^check "HTTP failed"
  |json/parse |^check "Bad JSON"
}
```

Breaking a forever:
```
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
sum { }                  ; 0
map { } { + 1 }          ; { }
filter { } { > 0 }       ; { }
first { }                ; error
reduce { } 'a { + a }    ; error
fold { } 'a 0 { + a }    ; 0
```

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
