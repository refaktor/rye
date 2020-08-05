# nekozito

> ⚠️ This is experimental project from [refaktor], looking into new language concepts and the depths of Go. ⚠️

## Development and experimentation

```bash

# Builds the "rye" interpreter
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

go build -x -o rye .

# Executable
./rye 
```

## Installing on OSX

This code-base relies on GTK3. So make sure your machine has it.

```bash
brew install pkg-config gtk+3 adwaita-icon-theme
```

More [instructions here](https://www.gtk.org/docs/installations/macos/).

## Author

- [Janko Metelko][refaktor] - `<janko.itm@gmail.com>` - [Facebook Group](https://www.facebook.com/groups/866313463771373)

[refaktor]: https://github.com/refaktor
[otobrglez]: https://github.com/otobrglez

