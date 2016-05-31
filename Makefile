.PHONY: all testdeps checkfmt lint test build run vet checkswagger swagger runswagger

all: test

testdeps:
	go get -d -t ./...

checkfmt: testdeps
	@export output="$$(gofmt -s -l .)" && \
		[ -n "$${output}" ] && \
		echo "Unformatted files:" && \
		echo && echo "$${output}" && \
		echo && echo "Please fix them using 'gofmt -s -w .'" && \
		export status=1; exit $${status:-0}

deadcode:
	go get github.com/remyoudompheng/go-misc/deadcode
	go list ./... | sed -e "s;github.com/nytm/video-transcoding-api;.;" | xargs deadcode

lint: testdeps
	go get github.com/golang/lint/golint
	@for file in $$(git ls-files '*.go'); do \
		export output="$$(golint $${file})"; \
		[ -n "$${output}" ] && echo "$${output}" && export status=1; \
	done; \
	exit $${status:-0}

test: checkfmt lint vet deadcode checkswagger
	go test ./...

build:
	go build

run: build
	./video-transcoding-api -config config.json

vet: testdeps
	go vet ./...

swagger:
	go get github.com/go-swagger/go-swagger/cmd/swagger
	swagger generate spec -o swagger.json

checkswagger:
	swagger validate swagger.json

runswagger:
	go run swagger-ui-server/main.go
