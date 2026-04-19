# Stage 1: Build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/vendel-smpp-agent -ldflags="-s -w" .

# Stage 2: Runtime
FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/vendel-smpp-agent /app/vendel-smpp-agent
ENTRYPOINT ["/app/vendel-smpp-agent"]
