FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git sqlite-dev bash gcc libc-dev

WORKDIR /go/src/rye

COPY . .

RUN mkdir -p /out && \
    go build -x -tags "b_tiny" -o /out/rye .

FROM alpine AS bin
COPY --from=builder /out/rye /
ENTRYPOINT ["/rye"]
