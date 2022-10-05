.PHONY: default

default: build

gen:
	sqlc generate

build: gen
	go build -o app cmd/main.go

test:
	go test -cover ./...

run: build
	./app

clean:
	rm app
	rm -r postgres/gen/**
