# Build stage
FROM --platform=linux/amd64 golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY internal ./internal
COPY services/payment-service ./services/payment-service

# Build the application
WORKDIR /app/services/payment-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main .

# Runtime stage
FROM --platform=linux/amd64 alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/services/payment-service/main .

EXPOSE 8082

CMD ["./main"]

