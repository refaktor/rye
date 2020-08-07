FROM golang:1.14-alpine AS builder

RUN apk add --no-cache git gtk+3.0 gtk+3.0-dev sqlite-dev bash gcc libc-dev

WORKDIR /go/src/rye

COPY . .

RUN go get -d -v -insecure \
    github.com/pkg/profile \
    github.com/yhirose/go-peg \
    github.com/labstack/echo/middleware \
    github.com/labstack/echo-contrib/session \
    github.com/gotk3/gotk3/gtk \
    github.com/lib/pq \
    github.com/mattn/go-sqlite3 \
    github.com/nats-io/nats.go \
    github.com/shirou/gopsutil/mem \
    github.com/tobgu/qframe \
    github.com/peterh/liner

RUN mkdir -p /out && \
    go build -x -o /out/rye .

FROM alpine AS bin
COPY --from=builder /out/rye /
ENTRYPOINT ["/rye"]
