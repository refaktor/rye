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

