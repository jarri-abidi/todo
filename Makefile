.PHONY: default

default: build

build:
	go build -o todolist

test:
	go test -cover ./...

run: build
	./todolist
	