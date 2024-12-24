ENV ?= development

.PHONY : run

run: build
	ENV=${ENV} ./bin/api

build:
	go mod tidy
	go build -o ./bin/ ./cmd/api/

test:
	go test -cover -v $(shell go list ./... | grep -vE "cmd|config")

migrate.create:
	migrate create -seq -ext .sql -dir ./migrations ${NAME}

migrate.up:
	migrate -path ./migrations -database ${DB_CONN} up ${STEP}

migrate.down:
	migrate -path ./migrations -database ${DB_CONN} down ${STEP}

migrate.force:
	migrate -path ./migrations -database ${DB_CONN} force ${VERSION}

migrate.drop:
	migrate -path ./migrations -database ${DB_CONN} drop

migrate.version:
	migrate -path ./migrations -database ${DB_CONN} version
