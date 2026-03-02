# Stage 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/ender-modem-agent -ldflags="-s -w" .

# Stage 2: Runtime
FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/ender-modem-agent /app/ender-modem-agent
ENTRYPOINT ["/app/ender-modem-agent"]
