# Stage 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/ender-modem-agent -ldflags="-s -w" .

# Stage 2: Runtime
FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/ender-modem-agent /app/ender-modem-agent
ENTRYPOINT ["/app/ender-modem-agent"]
