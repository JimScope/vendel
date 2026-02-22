# Stage 1: Build frontend
FROM node:24-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.23-alpine AS go-builder
WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/ender -ldflags="-s -w" .

# Stage 3: Runtime
FROM alpine:3.21
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=go-builder /app/ender /app/ender
COPY --from=frontend-builder /app/frontend/dist /app/pb_public

EXPOSE 8090

CMD ["/app/ender", "serve", "--http=0.0.0.0:8090"]
