
# Rye: gentle introduction trough Python examples #3

# Lists

list_of_ints = [1, 2, 3, 4, 5]

mixed = [1, "Emily", "emily@example.com"]

# Length

len(mixed)


sum(list_of_ints)
max(list_of_ints)

# Retrieving elements or slices

mixed[0]
mixed[1:0]
mixed[0:5]
mixed[:-3]

# Python generally uses imperative operations that modify the list

# Adding to list - appends or inserts a value to existing list

list_of_ints.append(6)
list_of_ints.insert(0, 99)

# Modifying specific value of a list

mixed[1] = "Emily Jones"

# doubling all values
res = []
for it in list_of_ints:
    res.append(it * 2)                   # or

map(lambda x: x * 2, list_of_ints)

# removing all odd values

for i in list_of_ints[:]:
    if i % 2 != 0:
        list_of_ints.remove(i)           # or

[x for x in l if x % 2 == 0]             # or

filter(lambda x: x % 2 == 0, list_of_ints)

    

# visit reddit.com/r/ryelang for other parts



