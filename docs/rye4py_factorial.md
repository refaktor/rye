## Rye for Python programmers - factorial

Challenge: *Print first 10 factorial numbers*

First we make a more "classic" looking implementation:

```red
factorial: fn { n } { 
  either n > 0               ;either is like if/else
    { n * factorial n - 1 } 
    { 1 } 
}

for range 1 10 { :n          ; :n will make more sense later
  print factorial n
}
;1
;2
;6
;24
;120
;720
;5040
;40320
;362880
;3628800
```
For better of just for show, let's enhance on the Rye-isms ..

```red
; rye doesn't care about newlines
; function accepting one arg can be defined by a pipe function
; either function can be called as a pipe-word
factorial: pipe { :n > 0 |either { n * factorial n - 1 } { 1 } }

; rye has some weird idea of *returning functions*, they start with ^
factorial: pipe { :n > 0 |^if { n * factorial n - 1 } 1 }

; injected blocks are usefull again
; op-words are similar to pipe-words, but a little different
loop { .factorial .print }
; 1
; 2
; 6
; 24
; 120
; 720
; 5040
; 40320
; 362880
; 3628800
```

I found this nice iterative example of Factorial in Python:

```python
def factorial(n):
    fact = 1
    for i in range(1, n + 1):
        fact *= i
    return fact
    
print(factorial(12))
# 479001600
```
We can mimmic it with Rye ...

```red
factorial: fn { n } { 
  fact: 1 
  for range 1 n { :i 
    fact: fact * i 
  } 
}
print factorial 12
; 479001600
```
If I again try to use more of Rye-s features ...

```red
; btw, the comma is an optional expression guard 
factorial: fn { n } { fact: 1 , loop n { * fact :fact } fact }

; we can do with less
factorial: pipe { fact: 1 , .loop { * fact :fact } }

factorial 12 |print
; 479001600
```
*I am sure Pythonistas can use list comprehensions and come up with some shorter version too.*

