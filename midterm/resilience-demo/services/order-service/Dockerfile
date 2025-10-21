# Build stage
FROM --platform=linux/amd64 golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY internal ./internal
COPY services/order-service ./services/order-service

# Build the application
WORKDIR /app/services/order-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main .

# Runtime stage
FROM --platform=linux/amd64 alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/services/order-service/main .

EXPOSE 8080

CMD ["./main"]

