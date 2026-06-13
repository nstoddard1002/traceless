# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod ./
# RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /traceless ./cmd/server/main.go

# Stage 2: Final lightweight image
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /

# Copy binary from builder
COPY --from=builder /traceless /traceless
# Copy static files
COPY --from=builder /app/web /web

EXPOSE 8080

ENTRYPOINT ["/traceless"]
