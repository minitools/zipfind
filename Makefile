

all:
	go get -v
	go build -v

test:
	go test -v ./...

.PHONY: all test