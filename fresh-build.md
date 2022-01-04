# How to build Rye on a fresh Ubuntu

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

    go get github.com/yhirose/go-peg # PEG parser (rye loader)
    go get github.com/peterh/liner   # library for REPL
    go get golang.org/x/net/html     # for html parsin - will probably remove for b_tiny
    go get github.com/pkg/profile    # for runtime profiling - will probably remove for b_tiny

Build the tiny version of Rye:

    go build -tags "b_tiny"

Run the rye file:

    ./rye hello.rye

Run the REPL

    ./rye

The language is moving towards the end of design stage. Current implementation was mostly a vehicle to
test and explore various language design ideas. It will get more unified, solidified and maybe even useful 
in 2022! :)






