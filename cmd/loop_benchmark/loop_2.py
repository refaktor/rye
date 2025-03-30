import time

start_time = time.time()
for _ in range(1, 10000000):
    _ = "text".upper()
end_time = time.time()

print(f"Execution time: {end_time - start_time} seconds")
