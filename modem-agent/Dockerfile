# Stage 1: Build
# Build context must be the repo root so agent-core/ is available via replace.
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY agent-core /src/agent-core
COPY modem-agent /src/modem-agent
WORKDIR /src/modem-agent
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/vendel-modem-agent -ldflags="-s -w" .

# Stage 2: Runtime
FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/vendel-modem-agent /app/vendel-modem-agent
ENTRYPOINT ["/app/vendel-modem-agent"]
