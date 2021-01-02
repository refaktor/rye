<b><a href="./TOUR_0.html">Some Practical Rye</a> > Do, If and Switch</b>

# Simple webserver

## Hello world

Blocks don't evaluate on their own. This enables almost all that follows below.

```rye
hello-world: fn { r w } { write w "Hello world!" }

new-server ":8080"
|handle "/" ?hello-world

```
