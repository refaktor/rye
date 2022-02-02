FROM golang:1.14-alpine AS builder

RUN apk add --no-cache git sqlite-dev bash gcc libc-dev

WORKDIR /go/src/rye

COPY . .

RUN go get -d -v -insecure \
	github.com/refaktor/go-peg \
	github.com/refaktor/liner \
	golang.org/x/net/html \
	github.com/pkg/profile \
	github.com/pkg/term

RUN mkdir -p /out && \
    go build -x -tags "b_tiny" -o /out/rye .

FROM alpine AS bin
COPY --from=builder /out/rye /
ENTRYPOINT ["/rye"]
