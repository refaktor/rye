import time

start_time = time.time()
for x in range(1, 10000000):
    x = 1 + 1
end_time = time.time()

print(f"Execution time: {end_time - start_time} seconds")
