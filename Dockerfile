# BUILDER
FROM golang:1.23.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api


# RUNNER
FROM debian:stable-slim AS runner
ARG TARGET_ENV="development"
COPY --from=builder /api /api
COPY ./.env.$TARGET_ENV ./.env.$TARGET_ENV
EXPOSE 8081
ENV PORT=8081
ENV ENV=$TARGET_ENV
ENTRYPOINT ["/api"]

