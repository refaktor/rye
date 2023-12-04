FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git sqlite-dev bash gcc libc-dev

WORKDIR /go/src/rye

COPY . .

RUN bash go-get-tiny.bash \
    mkdir -p /out && \
    go build -x -tags "b_tiny" -o /out/rye .

FROM alpine AS bin
COPY --from=builder /out/rye /
ENTRYPOINT ["/rye"]
