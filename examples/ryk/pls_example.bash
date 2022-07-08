$ # I have a CSV file this time
$ cat pls.csv
Name,Gender,Score,"Max Level"
Jane Austin,F,130,5
"James, J. Bent",M,129,5
Radth Daver,M,660,99
Roe Jogan,M,220,33
Lea Prince,F,1334,101
$ # And Ryk now supports CSVs too, and can skip a header
$ cat pls.csv | ryk --skip --csv '.print'
{ "Jane Austin" "F" 130 5 }
{ "James, J. Bent" "M" 129 5 }
{ "Radth Daver" "M" 660 99 }
{ "Roe Jogan" "M" 220 33 }
{ "Lea Prince" "F" 1334 101 }
$ # Awk has begin and end blocks and so does Ryk now
$ cat pls.csv | ryk --skip --csv --begin 'prn "PLAYERS:"' '-> 0 |+ "," |prn'
PLAYERS: Jane Austin, James, J. Bent, Radth Daver, Roe Jogan, Lea Prince, ⏎                                                                                                                                                                                                    $ # I want to know the average score of players
$ cat pls.csv | ryk --skip --csv '-> 2 |collect' --end '.avg |print'
494
$ # And the minimal and maximal levels of female players
$ cat pls.csv | ryk --skip --csv ':row -> 1 |= "F" |if { collect 3 <- row }' \
                    --end '.min |prn , .max |print'
5 101
$ # Now let's get the top 3 players displayed
$ cat pls.csv | ryk --skip --csv '-> 3 |prn , -> 0 |print' | sort -nr | head -n3
101 Lea Prince
99 Radth Daver
33 Roe Jogan
$ # Let's show their places too
$ # (btw: you can define custom Rye function in .ryk-preload)
$ cat .ryk-preload
quote: fn1 { .concat* $"$ |concat $"$ }
$ # Now we can use quote function in Ryk
$ cat pls.csv | ryk --skip --csv '-> 3 |prn , -> 0 |quote |print' | sort -nr | head -n3 \
              | ryk --begin 'place: 1' 'prn place .concat "." , -> 1 |print , place: inc place'
1. Lea Prince
2. Radth Daver
3. Roe Jogan
$ 
$ Thank you!
