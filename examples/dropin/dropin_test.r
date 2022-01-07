

print "## Hi, this is a drop-in test / demo ##"

a: 10

number: drop-in "Provide a number" { }

my-add: fn { a b } { drop-in "Inside my-add function" { } }

print "the result was: " + my-add 33 number

print "Bye!"

