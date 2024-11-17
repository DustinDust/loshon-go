ENV ?= development

.PHONY : run

run: build
	ENV=${ENV} ./bin/api

build:
	go mod tidy
	go build -o ./bin/ ./cmd/api/
