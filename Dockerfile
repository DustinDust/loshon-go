FROM golang:1.23.0 AS build-stage

WORKDIR .
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM gcr.io/distroless/base-debian11 AS build-release-stage

COPY --from=build-stage ./api ./api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["./api"]
