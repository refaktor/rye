# rye 🌾

> ⚠️ This is experimental programming language from [refaktor], looking into new (and borrowed) language ideas with interpreter in Go. ⚠️

You could peek at [NOTES.md](./NOTES.md), where concepts and ideas are born and logged. The language introduction documents are coming.

## Development and experimentation

Currently tested on Linux and Mac.

### Builds the rye interpreter

```bash
go get -v github.com/pkg/profile \
  github.com/yhirose/go-peg \
  github.com/mattn/go-sqlite3 \
 
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

## OSX tips

Ryefull relies on GTK3. So make sure your machine has it.

```bash
brew install pkg-config gtk+3 adwaita-icon-theme
```

More [instructions here](https://www.gtk.org/docs/installations/macos/).

## Building Docker image

```bash
docker build -t refaktor/rye -f .docker/Dockerfile .
```

> Currently broken, because rye depends on GTK3 and few 
> other things that are not part of original golang Docker image.  

## Author

- [Janko Metelko][refaktor] - `<janko.itm@gmail.com>`

## Resources and contact

- [Rye programming language - work in progress - Facebook Group](https://www.facebook.com/groups/866313463771373)

 [refaktor]: https://github.com/refaktor
 [otobrglez]: https://github.com/otobrglez

