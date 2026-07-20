import time

start_time = time.time()
for _ in range(1, 1000000):
    _ = 123
end_time = time.time()
print(f"inter: {end_time - start_time} seconds")

start_time = time.time()
for _ in range(1, 1000000):
    _ = "123"
end_time = time.time()
print(f"string: {end_time - start_time} seconds")

start_time = time.time()
for _ in range(1, 1000000):
    _ = [ 123 ]
end_time = time.time()
print(f"block: {end_time - start_time} seconds")

start_time = time.time()
for _ in range(1, 1000000):
    _ = 1 + 1 + 1
end_time = time.time()
print(f"builtins: {end_time - start_time} seconds")

start_time = time.time()
a = 1
b = "2"
c = [ 3 ]
for _ in range(1, 1000000):
    _ = a
    _ = b
    _ = c
end_time = time.time()
print(f"word-lookup: {end_time - start_time} seconds")

start_time = time.time()
a = 1
b = "2"
c = [ 3 ]
for _ in range(1, 1000000):
    a = b = c =123
end_time = time.time()
print(f"mow-word: {end_time - start_time} seconds")

start_time = time.time()
def a():
    return 123
    
for _ in range(1, 1000000):
    _ = a()
    _ = a()
    _ = a()
end_time = time.time()
print(f"fn-flat: {end_time - start_time} seconds")


start_time = time.time()
def a():
    return b()

def b():
    return c()

def c():
    return 123
    
for _ in range(1, 1000000):
    _ = a()
end_time = time.time()
print(f"fn-nest: {end_time - start_time} seconds")


# function env creation cost
def a(): pass
for _ in range(1000000): a()

# argument cost
def a3(x, y, z): return x
for _ in range(1000000): a3(1, 2, 3)

# recursion depth cost
def fib(n): return n if n < 2 else fib(n-1) + fib(n-2)
fib(25)


# closure cost
def make_counter():
    n = [0]
    def inc():
        n[0] += 1
        return n[0]
    return inc
c = make_counter()
for _ in range(1000000): c()

# branching cost
for _ in range(1000000): 1 if 1 > 0 else 2


# Python
def add_nums(x, y):
    return x + y

start_time = time.time()
for _ in range(1, 1000000):
    _ = add_nums(10, 20)
end_time = time.time()
print(f"fn-args: {end_time - start_time} seconds")

# Python
start_time = time.time()
for _ in range(1, 1000000):
    _ = 10 + 20 + 30 + 40
end_time = time.time()
print(f"math-chain: {end_time - start_time} seconds")

start_time = time.time()
for _ in range(1, 1000000):
    _ = [1, 2, 3]
end_time = time.time()
print(f"block-alloc: {end_time - start_time} seconds")

# Where python should win the most

d = {"key": 100}
start_time = time.time()
acc = 0
for _ in range(1, 1000000):
    acc += d["key"]
print(f"dict-lookup: {time.time() - start_time:.4f}s")

start_time = time.time()
for i in range(1, 1000000):
    _ = f"hello-{i}"
print(f"string-format: {time.time() - start_time:.4f}s")

start_time = time.time()
acc = 0
for i in range(1, 1000000):
    acc += i
print(f"math-accumulate: {time.time() - start_time:.4f}s")

# offload to c
data = list(range(1000000))
start = time.time()
total = sum(data)
print(f"sum: {time.time()-start}")

# string building

parts = [str(i) for i in range(100000)]
start = time.time()
result = "".join(parts)


# regex

import re
pattern = re.compile(r'\d+')
start = time.time()
for _ in range(1000000):
    pattern.match("12345")

# sets

s = set(range(1000000))
start = time.time()
for i in range(1000000):
    _ = i in s

# list comprehension (absolute killer in py)
[x * 2 for x in range(1_000_000)]
