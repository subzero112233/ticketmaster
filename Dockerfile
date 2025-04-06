# Stage 1: Build the Go application
FROM golang:1.22.5 AS builder

WORKDIR /app

COPY . /app

RUN api/generate_server.sh && \
    CGO_ENABLED=0 go build -o app cmd/native/main.go


# Stage 2: Create a minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/app .

ENTRYPOINT ["./app"]