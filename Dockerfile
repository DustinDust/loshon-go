# BUILDER
FROM golang:1.23.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api


# RUNNER
FROM debian:stable-slim AS runner
ARG TARGET_ENV
COPY --from=builder /api /api
COPY ./.env.$TARGET_ENV ./.env.$TARGET_ENV
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/api"]

