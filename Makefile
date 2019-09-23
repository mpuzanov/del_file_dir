.PHONY: build
build:
	go build -v 

.PHONY: run
run: 
	go run .

.PHONY: test
test:
	go test -v -race -timeout 30s ./...

.DEFAULT_GOAL := build
