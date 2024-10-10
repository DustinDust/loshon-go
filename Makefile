.PHONY : run

run: build
	./bin/api

build:
	go mod tidy
	go build -o ./bin/ ./cmd/api/