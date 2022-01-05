# Rye 🌾

Rye is **design/work in progress** programming language based on ideas from Rebol and flawored by
Factor, Bash shell and Golang. It currently features a WIP golang based interpreter and REPL.

## Rye at a glance 🌾

Rye is **homoiconic**, Rye's code is also Rye's data.

```clojure
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

```clojure
if-jim: fn { name code } { if name = "jim" code }

visitor: "jim"
if-jim visitor { print "Welcome in!" }
; prints: Welcome in!
```

Rye has no statements. Everything is an **expression**, returns 
something. Even asignment returns a value, so you can assign
inline. Either (an if/else like function) returns the result of the evaluated
block and can be used like an ternary operator.

```clojure
direction: 'in
full-name: join3 name: "Jane" " " surname: "Doe"  
print either direction = 'in { "Hi" +_ full-name } { "Bye bye" +_ name }
; outputs: Hi Jane Doe
```

Rye has **more syntax types** than your average language.
And it has generic methods for short function names. *Get* and *send*
below dispatch on the *kind* of first argument (http uri and an email address).

```clojure
email: jim@example.com
content: html->text get http://www.example.com
send email "title" content
; sends email with contents of the webpage
```

Rye has an option of forward code flow. It has a concept of 
**op and pipe words**. Every function can
be called as ordinary function or as op/pipe word. It also 
has a concept of **injected blocks** like *for* below.

```clojure
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

```clojure
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

```clojure
read-all %mydata.json |check { 404 "couldn't read the file" }
|parse-json |check { 500 "couldn't parse JSON" }
-> "score" |fix { 50 } |print1 "Your score: {}"
; outputs: Your score: 50
; if file is there and json is OK, but score field is missing
```

Rye has **multiple dialects**, specific interpreters of Rye values. 
Two examples of this are the validation and SQL dialects.

```clojure
dict { "name" "jane" surname: "boo" }
|validate { name: required score: optional 0 integer } |probe
// prints: #[ name: "jane" score: 0 ]

name: "James"
db: open sqlite://main.db
exec db { insert into pals ( name ) values ( ?name ) }
```

Rye has a concept of **kinds with validators and converters**.

```clojure
person: kind 'person { name: required string calc { .capitalize } }
user: kind user { user-name: required }
converter person user { user-name: :name }
dict { "name" "jameson" } >> person >> user |print
; prints: <Context (user): user-name: <String: Jameson>>
```

Rye's **scope/context** is a first class Rye value and by this very malleable.
One of the results of this are isolated contexts.

```clojure
iso: isolate { print: ?print plus: ?add }
do-in iso { 100 .plus 11 |print }
; prints 111
do-in iso { 2 .add 3 |print }
; Error: Word add not found
 ```
## Doing Y with Rye

Rye-s first focus is data (pre)processing, so some examples of that.

### read a webpage, save it to a file

```clojure
get https://www.google.com/search?q=ryelang
 |write-all %ryelang-results.html
```

### read and XML file and parse out specific information

rye has a SAX-based dialect

```clojure
"<people><person age='33'>Jim</person><person age='30'>Jane</person></people>" 
|string-reader |do-sxml { <person> [ .get-attr 0 |prn  ] _ [ .print ] }
; prints:
; 33 jim
; 30 jane
```



## More info

There is a [simple website](https://refaktor.github.io/rye/) being made, and a [blog](http://ryelang.blogspot.com/).

There is also a very small [FB group](https://www.facebook.com/groups/866313463771373) you can join.

## Platforms

Currently, tested on Linux, Mac and Docker.

### Building a minimal Rye

```bash
export GO111MODULE=auto

go get github.com/yhirose/go-peg # PEG parser (rye loader)
go get github.com/peterh/liner   # library for REPL
go get golang.org/x/net/html     # for html parsin - will probably remove for b_tiny
go get github.com/pkg/profile    # for runtime profiling - will probably remove for b_tiny

go build -tags "b_tiny"

# Executable
./rye 
```
More information on https://github.com/refaktor/rye/blob/main/fresh-build.md

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

- [Janko][refaktor] - `<janko.itm@gmail.com>`

## Contact

 [refaktor]: https://github.com/refaktor
 [otobrglez]: https://github.com/otobrglez

