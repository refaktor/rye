<b><a href="./TOUR_0.html">Some Practical Rye</a> > Simple webserver</b>

# Simple webserver

## Http server

Blocks don't evaluate on their own. This enables almost all that follows below.

```rye
new-server ":8080"
|handle "/" "Hello from Rye!"
|handle "/fn" fn { r w } { write w "Hello from Rye function!" }
```

## Websockets

```rye
new-server ":8080"
|handle-ws "/echo" fn { s ctx } { forever { read s ctx :m , write s ctx m } } 
|handle-ws "/ping" fn { s ctx } { forever { reas s ctx :m = "Ping" |if { write s ctx "Pong" } } }
```
