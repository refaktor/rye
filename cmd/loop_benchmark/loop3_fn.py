import time

def a():
    x = 1
    return x

start_time = time.time()
for _ in range(1, 10000000):
    a
end_time = time.time()

print(f"Execution time: {end_time - start_time} seconds")
