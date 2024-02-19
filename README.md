# Rye language 🌾

[![Build and Test](https://github.com/refaktor/rye/actions/workflows/build.yml/badge.svg)](https://github.com/refaktor/rye/actions/workflows/build.yml)
[![golangci-lint](https://github.com/refaktor/rye/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/refaktor/rye/actions/workflows/golangci-lint.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/refaktor/rye/badge)](https://securityscorecards.dev/viewer/?uri=github.com/refaktor/rye)
[![Go Reference](https://pkg.go.dev/badge/github.com/refaktor/rye.svg)](https://pkg.go.dev/github.com/refaktor/rye)
[![Go Report Card](https://goreportcard.com/badge/github.com/refaktor/rye)](https://goreportcard.com/report/github.com/refaktor/rye)
[![GitHub Release](https://img.shields.io/github/release/refaktor/rye.svg?style=flat)](https://github.com/refaktor/rye/releases/latest)

- [Rye language 🌾](#rye-language-)
  - [What is Rye](#what-is-rye)
  - [Status: Alpha](#status-alpha)
  - [Language overview](#language-overview)
    - [Few oneliners](#few-one-liners)
    - [Meet Rye](#meet-rye)
    - [Examples](#examples)
    - [Rye vs. Python](#rye-vs-python)
  - [Modules](#modules)
    - [Base](#base)
    - [Contrib](#contrib)
  -  [Rye-front project](#rye-front-project)
  - [Follow development](#follow-development)
    - [Rye blog](#rye-blog)
    - [Ryelang reddit](#ryelang-reddit)
    - [Github](#github)
  - [Getting Rye](#getting-rye)
    - [Binaries](#binaries)
    - [Docker images](#docker-images)
      - [Binary Docker image](#binary-docker-image)
      - [Dev Docker image](#dev-docker-image)
    - [Building Rye from source](#building-rye-from-source)
      - [Build Rye with specific modules](#build-rye-with-specific-modules)
      - [Build WASM version](#build-wasm-version)
      - [Tests and function reference](#tests-and-function-reference)
  - [Editor support](#editor-support)
  - [Related links](#related-links)
  - [Questions, contact](#questions-contact)

*visit **[ryelang.org](https://ryelang.org/)**, **[our blog](https://ryelang.org/blog/)** or join our **[reddit group](https://reddit.com/r/ryelang/)** for latest examples and development updates*

## What is Rye

Rye is a high level, dynamic **programming language** based on ideas from **Rebol**, flavored by
Factor, Linux shells and Golang. It's still an experiment in language design, but it should slowly become more and
more useful in real world.

It features a Golang based interpreter and console and could also be seen as (modest) **Go's scripting companion** as
Go's libraries are quite easy to integrate, and Rye can be embedded into Go programs as a scripting or config language.

I believe that as language becomes higher level it starts touching the user interface boundary, besides being a language
we have great emphasis on **interactive use** (Rye shell) where we will explore that.

## Status: Alpha

Core ideas of the language are formed. Most experimenting, at least at this stage, is done.
Right now, focus is on making the core and runtime useful for anyone who might decide to try it.

This means we are improving the Rye console, documentation and improving the runtime and core functions.

## Language overview

Rye is **homoiconic**, it has **no keywords** or special forms (everything is a function call, everything is a value), everything returns something (is an expression),
has more **syntax types** than your usual language, functions are **first class** citizens, so are blocks of code and scopes (contexts). It has multiple **dialects** (specific interpreters).

Although it seems contrary to each other Rye tries to be very **flexible** but also **safer** where possible. For example, it doesn't even have a syntax for changing state directly 
outside current context (parent or sub). It separates between pure and impure functions, while most of builtins are pure. Validation dialect is part of its core, so input validation is easy and 
distinguishable/explicit, not sprinkled around other code. Few functions that change state in place end with "!" (and usually don't need to be used). Functions never return null, they return result
or a specific failure (which is a Rye value too, and you can handle on-specific-spot). 

### Few one-liners

```red
print "Hello World"

"Hello World" .replace "World" "Mars" |print
; prints: Hello Mars

"12 8 12 16 8 6" .load .unique .sum
; returns: 42

{ "Anne" "Joan" "Adam" } |filter { .first = "A" } |for { .print } 
; prints:
; Anne
; Adam

fac: fn { x } { either x = 1 { 1 } { x * fac x - 1 } }
; function that calculates factorial

range 1 10 |map { .fac } |print\csv
; prints: 1,2,6,24,120,720,5040,40320,362880,3628800

kind: "admin"
open sqlite://data.db |query { select * from user where kind = ?kind }
; returns: Spreadsheet of admins

read %name.txt |fix { "Anonymous" } |post* https://example.com/postname 'text
; makes HTTP post of the name read from a file, or "Anonymous" if file failed to be read
```

### Meet Rye

Visit this set of pages to find out more (work in progress):

  * [Meet Rye](https://ryelang.org/meet_rye/)
 
### Examples

These pages are littered with examples. You can find them on this **README** page, on our **blog**, on **Introductions**, but also:

  * [Examples folder](./examples/) - unorganized so far, will change
  * [Solutions to simple compression puzzle](https://github.com/otobrglez/compression-puzzle/tree/master/src/rye)

### Rye vs. Python

Python is the *lingua franca* of dynamic programming languages, so comparing examples in Python and Rye can be helpful to some:

  * [Less variables, more flows example vs Python](https://ryelang.blogspot.com/2021/11/less-variables-more-flows-example-vs.html)
  * [Simple compression puzzle - from Python to Rye solution](https://github.com/otobrglez/compression-puzzle/blob/master/src/rye/compress_jm_rec_steps.rye)
  
## Modules

The author of Factor once said that at the end *it's not about the language, but the libraries*. I can only agree, adding *libraries, and distribution*. 

### Base
  * Core builtin functions ⭐⭐⭐  🧪~80%
  * Bcrypt - password hashing
  * Bson - binary (j)son
  * Crypto - cryptographic functions ⭐
  * Email - email generation and parsing
  * Html - html parsing
  * Http - http servers and clients ⭐⭐
  * IO (!) - can be excluded at build time ⭐⭐ 
  * Json - json parsing ⭐⭐ 
  * Mysql - database ⭐
  * Postgresql - database ⭐
  * Psutil - linux process management
  * Regexp - regular expressions ⭐ 🧪~50%
  * Smtpd - smtp server (receiver)
  * Sqlite - database ⭐⭐
  * Sxml - sax XML like streaming dialect
  * Validation - validation dialect ⭐⭐ 🧪~50% 
   
### Contrib
  * Amazon AWS
  * Bleve full text search 
  * Cayley graph database
  * OpenAI - OpenAI API
  * Postmark - email sending service
  * Telegram bot - telegram bots

legend: ⭐ priority , 🧪 tests

## Rye-front project

If you are interested in "fontend" / desktop technologies check out separate project that works on extending Rye lanuage with GUI, Game engine and a Webview. It 
integrates these cool Go libraries:

  * Fyne - Cross platform Material design inspired GUI framework (desktop and mobile) 
  * Ebitengine - 2d game engine (desktop, mobile and web)
  * Webview - Webview GUI

**[Visit Rye-front repo](https://github.com/refaktor/rye-front)**
    
## Follow development

### Rye blog

For most up-to date information on the language, and its development, visit our **[old](http://ryelang.blogspot.com/)** and **[new blog](http://ryelang.org/blog)**.

### Ryelang reddit

This is another place for updates and also potential discussions. You are welcome to join **[our reddit group](https://reddit.com/r/ryelang/)**. 

### Github

If code speaks to you, our Github page is the central location for all things Rye. You are welcome to collaborate, post Issues or PR-s, there are tons of things to do and improve :)

## Getting Rye

Rye is developed on Linux, but has also been compiled on macOS, Docker and as WASM module. If you need additional architecture or OS, post an Issue.

### Binaries

You can find precompiled Binaries for **Linux** and **macOS** under [Releases](https://github.com/refaktor/rye/releases).

Docker images are published under [Packages](https://github.com/refaktor/rye/pkgs/container/rye).

### Docker images

#### Binary Docker image

This image includes Linux, Rye binary ready for use and Emacs-nox editor.

Docker image: **[ghcr.io/refaktor/rye](https://github.com/refaktor/rye/pkgs/container/rye)**

Run it via:

```bash
docker run -ti ghcr.io/refaktor/rye
```

#### Dev Docker image

The repository comes with a local [Docker image](./.docker/Dockerfile) that builds rye and allows you to do so.

```bash
docker build -t refaktor/rye -f .docker/Dockerfile .
```

Run 🏃‍♂️ the rye REPL with:

```bash
docker run -ti refaktor/rye
```

### Building Rye from source

Use official documentation or lines below to install Golang 1.19.3 https://go.dev/doc/install (at the time of writing):

    wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    go version
    
Clone the main branch from the Rye repository:

    git clone https://github.com/refaktor/rye.git && cd rye

Build the tiny version of Rye:

    go build -tags "b_tiny"

Run the rye file:

    ./rye hello.rye

Run the Rye Console

    ./rye

Install build-essential if you don't already have it, for packages that require cgo (like sqlite):

    sudo apt install build-essential

#### Build Rye with specific modules

Rye has many bindings, that you can determine at build time, so (currently) you get a specific Rye binary for your specific project. This is an example of a build with many bindings. 
I've been working on a way to make this more elegant and systematic, but the manual version is:

    go build -tags "b_tiny,b_sqlite,b_http,b_sql,b_postgres,b_openai,b_bson,b_crypto,b_smtpd,b_mail,b_postmark,b_bcrypt,b_telegram"
	
#### Build WASM version

Rye can also work inside your browser or any other WASM container. I will add examples, html pages and info about it, but to build it:

    GOOS=js GOARCH=wasm go build -tags "b_tiny" -o wasm/rye.wasm main_wasm.go
	
Run the demo server for testing WASM version

    bin/rye serve_wasm.rye
	
Then visit http://localhost:8085 or http://localhost:8085/ryeshell/

#### Tests and function reference

Run the Rye code tests.

    cd tests
	../bin/rye main.rye test
	
Build the function reference out of tests:

    cd tests
	../bin/rye main.rye doc

## Editor support

Rye has Syntax highlighting for Emacs and VS Code. For VS Code just search for **ryelang** in the Extension marketplace. For Emacs it will be published soon on github. 

## Related links

  [**Rebol**](http://www.rebol.com) - Rebol's author Carl Sassenrath invented or combined together 90% of concepts that Rye builds upon.
  
  [Factor](https://factorcode.org/) - Factor from Slava Pestov taught me new fluidity and that variables are *no-good*, but stack shuffle words are even worse ;)
  
  [Red](https://red-lang.org) - Another language inspired by Rebol from well known Rebol developer  DocKimbel and his colleagues. A concrete endeavor, with its low level language, compiler, GUI, ...
  
  [Oldes' Rebol 3](https://oldes.github.io/Rebol3/) - Rebol3 fork maintained by Oldes (from Amanita Design), tries to resolve issues without unnecessarily changing the language itself.
  
  [Arturo](https://arturo-lang.io/) - Another unique language that builds on Rebol's core ideas.

  [Charm](https://github.com/tim-hardcastle/Charm) - Not related to Rebol, but an interesting Go based language with some similar runtime ideas and challenges.
  
  [Ren-c](https://github.com/metaeducation/ren-c) - Rebol 3 fork maintained by HostileFork, more liberal with changes to the language. 
  
 ## Questions, contact
  
You can post an **[Issue](https://github.com/refaktor/rye/issues)** visit github **[Discussions](https://github.com/refaktor/rye/discussions)** or contact me through <a href="mailto:janko .itm+rye[ ]gmail .com">gmail</a> or <a href="https://twitter.com/refaktor">twitter</a>.</p>

