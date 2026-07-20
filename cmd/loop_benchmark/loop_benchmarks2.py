import time
import math

start_time = time.time()
for _ in range(1, 1000000):
    _ = 123
end_time = time.time()
print(f"integer: " + str(math.ceil((end_time - start_time) * 1000)))

start_time = time.time()
for _ in range(1, 1000000):
    _ = "123"
end_time = time.time()
print(f"string: " + str(math.ceil((end_time - start_time) * 1000)))

start_time = time.time()
for _ in range(1, 1000000):
    _ = [123]
end_time = time.time()
print(f"block: " + str(math.ceil((end_time - start_time) * 1000)))

start_time = time.time()
for _ in range(1, 1000000):
    _ = 1 + 1 + 1
end_time = time.time()
print(f"builtins: " + str(math.ceil((end_time - start_time) * 1000)))

a = 1
b = "2"
c = [3]
start_time = time.time()
for _ in range(1, 1000000):
    _ = a
    _ = b
    _ = c
end_time = time.time()
print(f"word-lookup: " + str(math.ceil((end_time - start_time) * 1000)))

a = 1
b = "2"
c = [3]
start_time = time.time()
for _ in range(1, 1000000):
    a = b = c = 123
end_time = time.time()
print(f"mod-word: " + str(math.ceil((end_time - start_time) * 1000)))

def a():
    return 123

start_time = time.time()
for _ in range(1, 1000000):
    _ = a()
    _ = a()
    _ = a()
end_time = time.time()
print(f"fn-flat: " + str(math.ceil((end_time - start_time) * 1000)))

def a():
    return b()
def b():
    return c()
def c():
    return 123

start_time = time.time()
for _ in range(1, 1000000):
    _ = a()
end_time = time.time()
print(f"fn-nest: " + str(math.ceil((end_time - start_time) * 1000)))

def a():
    pass

start_time = time.time()
for _ in range(1000000):
    a()
end_time = time.time()
print(f"fn-env: " + str(math.ceil((end_time - start_time) * 1000)))

def a3(x, y, z):
    return x

start_time = time.time()
for _ in range(1000000):
    a3(1, 2, 3)
end_time = time.time()
print(f"fn-args3: " + str(math.ceil((end_time - start_time) * 1000)))

def fib(n):
    return n if n < 2 else fib(n - 1) + fib(n - 2)

start_time = time.time()
fib(25)
end_time = time.time()
print(f"recursion-fib25: " + str(math.ceil((end_time - start_time) * 1000)))

def make_counter():
    n = [0]
    def inc():
        n[0] += 1
        return n[0]
    return inc

c = make_counter()
start_time = time.time()
for _ in range(1000000):
    c()
end_time = time.time()
print(f"closure: " + str(math.ceil((end_time - start_time) * 1000)))

start_time = time.time()
for _ in range(1000000):
    _ = 1 if 1 > 0 else 2
end_time = time.time()
print(f"branching: " + str(math.ceil((end_time - start_time) * 1000)))

def add_nums(x, y):
    return x + y

start_time = time.time()
for _ in range(1, 1000000):
    _ = add_nums(10, 20)
end_time = time.time()
print(f"fn-args: " + str(math.ceil((end_time - start_time) * 1000)))

a = 0
start_time = time.time()
for _ in range(1, 1000000):
    a = 123
end_time = time.time()
print(f"var-assign: " + str(math.ceil((end_time - start_time) * 1000)))

start_time = time.time()
for _ in range(1, 1000000):
    _ = 10 + 20 + 30 + 40
end_time = time.time()
print(f"math-chain: " + str(math.ceil((end_time - start_time) * 1000)))

start_time = time.time()
for _ in range(1, 1000000):
    _ = [1, 2, 3]
end_time = time.time()
print(f"block-alloc: " + str(math.ceil((end_time - start_time) * 1000)))

d = {"key": 100}
acc = 0
start_time = time.time()
for _ in range(1, 1000000):
    acc += d["key"]
end_time = time.time()
print(f"dict-lookup: " + str(math.ceil((end_time - start_time) * 1000)))

start_time = time.time()
for i in range(1, 1000000):
    _ = f"hello-{i}"
end_time = time.time()
print(f"string-format: " + str(math.ceil((end_time - start_time) * 1000)))

acc = 0
start_time = time.time()
for i in range(1, 1000000):
    acc += i
end_time = time.time()
print(f"math-accumulate: " + str(math.ceil((end_time - start_time) * 1000)))

data = list(range(1000000))
start_time = time.time()
total = sum(data)
end_time = time.time()
print(f"offload-sum: " + str(math.ceil((end_time - start_time) * 1000)))

parts = [str(i) for i in range(100000)]
start_time = time.time()
result = "".join(parts)
end_time = time.time()
print(f"string-build: " + str(math.ceil((end_time - start_time) * 1000)))

import re
pattern = re.compile(r'\d+')
start_time = time.time()
for _ in range(1000000):
    pattern.match("12345")
end_time = time.time()
print(f"regex-match: " + str(math.ceil((end_time - start_time) * 1000)))

s = set(range(1000000))
start_time = time.time()
for i in range(1000000):
    _ = i in s
end_time = time.time()
print(f"set-lookup: " + str(math.ceil((end_time - start_time) * 1000)))

start_time = time.time()
_ = [x * 2 for x in range(1000000)]
end_time = time.time()
print(f"list-comprehension: " + str(math.ceil((end_time - start_time) * 1000)))

