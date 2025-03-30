
## BENCHMARKS

### MAIN

*Doing simple loop 1million to get some basic feel for how fast certain things are*
DO mode
integer: 123 79
string: 124 48
block: 123  42
builtins (plus plus): 3 287
word lookup: 3  138
mod-word: 123 98
fn-flat: 123 598
fn-nest: 123 587
RYE0 mode
integer: 123 40
string: 124 29
block: 123  74
builtins (plus plus): 3 215
word lookup: 3  131
mod-word: 123 93
fn-flat: 123 537
fn-nest: 123 558

### 90prep

*Doing simple loop 1million to get some basic feel for how fast certain things are*
DO mode
integer: 123 37
string: 124 25
block: 123  28
builtins (plus plus): 3 243
word lookup: 3  137
mod-word: 123 108
fn-flat: 123 526
fn-nest: 123 553
RYE0 mode
integer: *
*
123 19
string: *
*
124 23
block: *
*
123  65
builtins (plus plus): *
*
*
*
3 75
word lookup: *
*
3  118
mod-word: *
*
123 68
fn-flat: *
*
123 591
fn-nest: *
*
123 619

### Optim

*Doing simple loop 1million to get some basic feel for how fast certain things are*
DO mode
integer: 123 47
string: 124 28
block: 123  28
builtins (plus plus): 3 231
word lookup: 3  131
mod-word: 123 81
fn-flat: 123 530
fn-nest: 123 2465
RYE0 mode
integer: 123 22
string: 124 21
block: 123  68
builtins (plus plus): 3 81
word lookup: 3  151
mod-word: 123 74
fn-flat: 123 274
fn-nest: 123 317


## OBSERVATIONS

Do mode is faster on 90prep and optim, except nested fn on Optim are 5x worse - 50% better

Rye0 builtin is the best at 90prep 5% , second Optim
Rye0 fn is the best at Optim , second prep , 50% !! razlike


## CONCLUSION

DO - 90prep is the best ... stufy differences to main and implement them except bug at secondArdLeft (*)

RYE0 
- builtin a little better on  90prep than Optim, compare and make the cleanest option
- fn optimisations take from Optim
