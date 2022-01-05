.PHONY: clean install-dependencies
.MAIN: rye

rye: install-dependencies
	go build -tags "b_tiny"

clean:
	rm -rf rye

install-dependencies:
	go get -v \
		github.com/yhirose/go-peg  \
		github.com/peterh/liner \
		golang.org/x/net/html \
		github.com/pkg/profile
