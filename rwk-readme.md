# RWK - Rye meets Awk [alpha]

## Example

$ cat stars.csv 
7,"James, James Bond"
9,Geralt of Rivia
12,Ellen Ripley

$ cat stars.csv | rye rwk --csv --begin 'print "STARS:"' '-> 1 |prns , -> 0 |produce "" { .concat "*" } |print' 
STARS:
James, James Bond *******
Geralt of Rivia *********
Ellen Ripley ************

