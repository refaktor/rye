# Rye language 🌾

   * [What is it](#what-is-it)
   * [Alpha status](#alpha-status)
   * [Some Rye specifics](#some-rye-specifics)
   * [Follow the blog](#find-the-blog)
   * [How to install / build](#how-to-install--build)
   * [Docker image](#docker-image)

## What is it

Rye is a human and information centric, high level, dynamic **programming language** based on ideas from **Rebol**, flavored by
Factor, Linux shells and Golang. It features a Golang based interpreter and REPL.

## Alpha status

Core ideas of the language are formed. Most experimenting, at least until version 1 is done.
Right now, my focus is on making the core and runtime more ready and friendly for a potential brave soul that wants to install it and 
dabble around a little.

That means I am improving the REPL, live documentation and stabilising core functions.

## Some Rye specifics

Rye is **homoiconic**, Rye's code is also Rye's data.

```red
data: { name "Jim" score 33 }
code: { print "Hello" +_ data/name }
if data/score > 25 code
; outputs: Hello Jim
print second data + ", " + second code
; outputs: Jim, Hello
```

Rye has **no keywords or special forms**. Ifs, Loops, even Function
definiton words are just functions. This means you can have more of them,
load them at library level and create your own.

```red
if-jim: fn { name code } { if name = "jim" code }

visitor: "jim"
if-jim visitor { print "Welcome in!" }
; prints: Welcome in!
```

Rye has no statements. Everything is an **expression**, returns 
something. Even asignment returns a value, so you can assign
inline. Either (an if/else like function) returns the result of the evaluated
block and can be used like an ternary operator.

```red
direction: 'in
full-name: join3 name: "Jane" " " surname: "Doe"  
print either direction = 'in { "Hi" +_ full-name } { "Bye bye" +_ name }
; outputs: Hi Jane Doe
```

Rye has **more syntax types** than your average language.
And it has generic methods for short function names. *Get* and *send*
below dispatch on the *kind* of first argument (http uri and an email address).

```red
email: jim@example.com
content: html->text get http://www.example.com
send email "title" content
; sends email with contents of the webpage
```

Rye has an option of forward code flow. It has a concept of 
**op and pipe words**. Every function can
be called as ordinary function or as op/pipe word. It also 
has a concept of **injected blocks** like *for* below.

```red
http://www.example.com/ .get .html->text :content
jim@example.com .send "title" content
; sends email with contents of the webpage
{ "Jane" "Jim" } |for { .prn }
; outputs: Jane Jim
get http://www.example.com/users.json
|parse-json |for { -> "name" |capitalize |print }
; outputs capitalized names of users
```

Rye has **higher order like functions**, but they come in what
would usually be special forms (here these are still just functions).

```red
nums: { 1 2 3 }
map nums { + 30 }
; returns { 31 32 33 }
filter nums { :x all { x > 1  x < 3 } }
; returns: { 2 }
strs: { "one" "two" "three" }
{ 3 1 2 3 } |filter { > 1 } |map { <-- strs } |for { .prn }
; prints: three two three
```

Rye has some different ideas about **failure handling**. This
is a complex subject, so look at other docs about it. Remember it's
all still experimental.

```red
read-all %mydata.json |check { 404 "couldn't read the file" }
|parse-json |check { 500 "couldn't parse JSON" }
-> "score" |fix { 50 } |print1 "Your score: {}"
; outputs: Your score: 50
; if file is there and json is OK, but score field is missing
```

Rye has **multiple dialects**, specific interpreters of Rye values. 
Two examples of this are the validation and SQL dialects.

```red
dict { "name" "jane" surname: "boo" }
|validate { name: required score: optional 0 integer } |probe
// prints: #[ name: "jane" score: 0 ]

name: "James"
db: open sqlite://main.db
exec db { insert into pals ( name ) values ( ?name ) }
```

Rye has a concept of **kinds with validators and converters**.

```red
person: kind 'person { name: required string calc { .capitalize } }
user: kind user { user-name: required }
converter person user { user-name: :name }
dict { "name" "jameson" } >> person >> user |print
; prints: <Context (user): user-name: <String: Jameson>>
```

Rye's **scope/context** is a first class Rye value and by this very malleable.
One of the results of this are isolated contexts.

```red
iso: isolate { print: ?print plus: ?add }
do-in iso { 100 .plus 11 |print }
; prints 111
do-in iso { 2 .add 3 |print }
; Error: Word add not found
 ```
 
## Follow the blog

For most up-to date information on the language and it's development visit out **[blog](http://ryelang.blogspot.com/)**

*There is a work in progress [website](https://refaktor.github.io/rye/).*

## How to install / build 

Rye is developed on Linux, but has also been compiled on Mac and Docker.

### How to build Rye on a fresh Ubuntu

Use official documentation to install latest Golang https://go.dev/doc/install

    cd /tmp
    wget https://go.dev/dl/go1.17.5.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.17.5.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin

Check Go version.

    go version
    go version go1.17.5 linux/amd64

Create the ~/go/src directory

    mkdir -p ~/go/src && cd ~/go/src

Clone the main branch from the Rye repository

    git clone https://github.com/refaktor/rye.git && cd rye

Rye doesn't use go modules yet, but "go get"

    export GO111MODULE=auto

Go get the dependencies for tiny build

    go get github.com/refaktor/go-peg  # PEG parser (rye loader)
    go get github.com/refaktor/liner   # library for REPL
    go get golang.org/x/net/html       # for html parsin - will probably remove for b_tiny
    go get github.com/pkg/profile      # for runtime profiling - will probably remove for b_tiny
    go get github.com/pkg/term         # 

Build the tiny version of Rye:

    go build -tags "b_tiny"

Run the rye file:

    ./rye hello.rye

Run the REPL

    ./rye

## Docker image

The repository comes with [Docker image](./.docker/Dockerfile) that is capable of building rye in its full, 
the final step however then just wraps executable so that the image remains small and nimble.

```bash
docker build -t refaktor/rye -f .docker/Dockerfile .
```

Run 🏃‍♂️ the rye REPL with:

```bash
docker run -ti refaktor/rye
```




