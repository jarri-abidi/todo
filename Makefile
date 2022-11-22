.PHONY: default

default: build

gen:
	sqlc generate

build: gen test
	go build -o app cmd/main.go

test:
	go test -cover ./...

run: build
	./app

clean:
	rm -r pkg/postgres/gen/**
