FROM golang:1.14

WORKDIR /src

COPY . .

RUN go get -d -v -insecure github.com/pkg/profile \
  github.com/yhirose/go-peg \
  github.com/labstack/echo/middleware \
  github.com/labstack/echo-contrib/session \
  github.com/gotk3/gotk3/gtk \
  github.com/lib/pq \
  github.com/mattn/go-sqlite3 \
  github.com/nats-io/nats.go \
  github.com/shirou/gopsutil/mem \
  github.com/tobgu/qframe

RUN go build -x -o rye

FROM scratch AS bin
COPY --from=build /out/example /
