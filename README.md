# Rye 🌾

Rye is work/design in progress programming language based on ideas from Rebol merged with some ideas
from factor, bash shell and golang. It currently features a WIP golang based interpreter.

# Rye at a glance 🌾

Rye is **homoiconic**, code is also Rye's data.

```bash
data: { name "Jim" score 33 }
code: { print "Hello" +_ data/name }
if data/score > 25 code
; outputs: Hello Jim
print second data + ", " + second code
; outputs: Jim, Hello
```

Rye has **no keywords or special forms**. Ifs, Loops, even Function
definiton words are just functions. This means you can load them at 
library level and create your own.

```bash
if-jim: fn { name code } { if name = "jim" code }

visitor: "jim"
if-jim visitor { print "Welcome in!" }
; prints: Welcome in!
```

Rye has no statements. Everything is an **expression**, returns 
something. Even asignment returns a value, so you can assign
inline. (Either is an if/else like function).

```bash
direction: 'in
full-name: join3 name: "Jane" " " surname: "Doe"  
print either direction = 'in { "Hi" +_ full-name } { "Bye bye" }
; outputs: Hi Jane Doe
```

Rye has more **syntax types** than your average language.
And it has generic methods for short function names. *Get* and *send*
below dispatch on the kind of first argument (http uri and an email address).

```bash
email: jim@example.com
content: html->text get http://www.example.com
send email "title" content
; sends email with contents of the webpage
```

Rye has an option of forward code flow. It has a concept of 
**op and pipe words**. Every function can
be called as ordinary function or as op/pipe word.

```bash
http://www.example.com/ .get .html->text :content
jim@example.com .send "title" content
; sends email with contents of the webpage

get http://www.example.com/users.json
|parse-json |for { -> "name" |capitalize |print }
; ouptuts capitalized names of users
```

Rye has **higher order like functions**, but they come in what
would usually be special forms (here these are just functions).

```bash
nums: { 1 2 3 }
map nums { + 30 }
; returns { 31 32 33 }
filter nums { < 3 }
; returns { 1 2 }
filter nums { :x all { x > 1  x < 3 } }
; returns: { 2 }
strs: { "one" "two" "three" }
{ 3 1 2 3 } |filter { > 1 } |map { <-- strs } |for { .print }
; prints: 
;  three
;  two
;  three
```



Rye has some different ideas about **failure handling**. This
is a complex subject, so look at other docs about it. Remember it's
all still experimental.

```bash
read-all %mydata.json |check { 404 "couldn't read the file" }
|parse-json |check { 500 "couldn't parse JSON" }
-> "score" |fix { 50 } |print1 "Your score: {}"
; outputs: Your score: 50
; if file is there and json is OK, but score field is missing
```


## Examples and more info

A [blog](https://ryelang.blogspot.com/) following development.

The language [introduction documents](https://refaktor.github.io/rye/INTRO_1.html) are work in progress.

There is also a [simple website](https://refaktor.github.io/rye/) being made.

There is also a very small [FB group](https://www.facebook.com/groups/866313463771373) you can join.

## Platforms

Currently, tested on Linux, Mac and Docker.

### Builds the rye interpreter

```bash
go get -v github.com/pkg/profile \
  github.com/yhirose/go-peg \
  github.com/mattn/go-sqlite3
 
go build

# Executable
./rye 
```

### Builds the rye interpreter with all current builtins

You can leave just the tags and install just the bindings you want.

```bash
go get -v github.com/pkg/profile \
  github.com/yhirose/go-peg \
  github.com/labstack/echo/middleware \
  github.com/labstack/echo-contrib/session \
  github.com/gotk3/gotk3/gtk \
  github.com/lib/pq \
  github.com/mattn/go-sqlite3 \
  github.com/nats-io/nats.go \
  github.com/shirou/gopsutil/mem \
  github.com/tobgu/qframe

go build -tags "b_echo b_gtk b_psql b_nats b_psutil b_qframe" -o ryefull

# Executable
./ryefull 
```

## Docker image

The repository comes with [Docker image](./docker/Dockerfile) that is capable of building rye in its full, 
the final step however then just wraps executable so that the image remains small and nimble.

```bash
docker build -t refaktor/rye -f .docker/Dockerfile .
```

Run 🏃‍♂️ the rye REPL with:

```bash
docker run -ti refaktor/rye
```

## OSX tips

Ryefull relies on GTK3. So make sure your machine has it.

```bash
brew install pkg-config gtk+3 adwaita-icon-theme
```

More [instructions here](https://www.gtk.org/docs/installations/macos/).

## Author

- [Janko Metelko][refaktor] - `<janko.itm@gmail.com>`

## Resources and contact

- [Rye programming language - work in progress - Facebook Group](https://www.facebook.com/groups/866313463771373)

 [refaktor]: https://github.com/refaktor
 [otobrglez]: https://github.com/otobrglez

