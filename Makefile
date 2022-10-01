.PHONY: default

default: build

build:
	go build -o app cmd/main.go

test:
	go test -cover ./...

run: build
	./app

clean:
	rm -r postgres/gen/**
