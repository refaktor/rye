# Fantastic cereal saves the day!

## Day one: poppie flowers

There was a dry day in the field. Our rye specimen Jimny was a little bored decided to count the poppy flowers on it's field.
He counted 23 of them to the west, 13 to the north and only 6 to the south. He quickly summed the numbers and got 42. 42 poppies ...

    { rye } 23 + 13 + 6
    < Integer: 42 >
    
He tried to strike conversation with a Cloje (the neigbour), but she gave him bunch of weird questions ...

"How many poppies are in the west and north together?"
"How many more poppies are there on the north than the south?"
"What about the east?"

Jimny needed to organize better

    { rye } west: 23  south: 13  north: 6
    { rye } print west + north
    29
    { rye } print south - north
    7

Jimny couldn't see the ones to the east, so he asked the his eastern neighbour Pyeter

    { rye } ask-pyeter: fn { } { 33 }
    { rye } east: ask-pyeter
    { rye } list { west east south north } |sum |print
    73

<a href="./INTRO_2.html" class="next">Second page</a>
