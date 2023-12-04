FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git sqlite-dev bash gcc libc-dev

WORKDIR /go/src/rye

COPY . .

ENV GO111MODULE=auto

RUN go get -d -v -insecure \
	github.com/refaktor/go-peg \
	github.com/refaktor/liner \
	golang.org/x/net/html \
	github.com/pkg/profile \
	github.com/pkg/term && \
	mkdir -p /out && \
    go build -x -tags "b_tiny" -o /out/rye .

FROM alpine AS bin
COPY --from=builder /out/rye /
ENTRYPOINT ["/rye"]
