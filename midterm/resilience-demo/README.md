# Resilience Patterns Demo

A comprehensive demonstration of resilience patterns (Fail Fast, Circuit Breaker, and Bulkhead) using a 3-microservice system in Go with the Gin web framework, Prometheus for metrics collection, and Grafana for visualization.

## 🏗️ System Architecture

```
Client → Order Service (8080) → Inventory Service (8081)
                               ↘ Payment Service (8082)
                               
All Services → Prometheus (9090) → Grafana (3000)
```

### Services

1. **Order Service (Port 8080)** - Main orchestrator service that implements all resilience patterns
2. **Inventory Service (Port 8081)** - Manages product inventory with chaos engineering capabilities
3. **Payment Service (Port 8082)** - Handles payment processing with chaos engineering capabilities
4. **Prometheus (Port 9090)** - Metrics collection and storage
5. **Grafana (Port 3000)** - Metrics visualization and dashboards

## 🎯 Resilience Patterns Implemented

### 1. Fail Fast Pattern
- **Input Validation**: Rejects invalid orders immediately
- **Timeouts**: All HTTP calls have 3-second timeouts
- **Early Failure Detection**: Returns errors immediately when preconditions aren't met

**Benefits**: Prevents wasted resources, faster feedback to clients

### 2. Circuit Breaker Pattern
- **Library**: `sony/gobreaker`
- **Configuration**:
  - Max Requests in Half-Open: 3
  - Failure Threshold: 60% failure rate with ≥3 requests
  - Timeout: 30 seconds before attempting half-open state
- **States**: Closed → Open → Half-Open → Closed

**Benefits**: Prevents cascading failures, allows services to recover

### 3. Bulkhead Pattern
- **Implementation**: Semaphore-based resource isolation
- **Capacity**: 10 concurrent requests per downstream service
- **Timeout**: 1 second to acquire resource

**Benefits**: Limits resource consumption, prevents thread exhaustion

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development)
- curl and jq (for demo scripts)

### Starting the System

```bash
# Make scripts executable
chmod +x scripts/*.sh

# Start all services
./scripts/start-all.sh
```

The script will:
1. Build all Docker images
2. Start all services
3. Wait for services to be ready
4. Display access URLs

### Access Points

- **Order Service**: http://localhost:8080
- **Inventory Service**: http://localhost:8081
- **Payment Service**: http://localhost:8082
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (Login: admin/admin)

## 📊 Grafana Dashboard

The pre-configured dashboard includes:

1. **Request Rate** - Requests per second for each service
2. **Circuit Breaker State** - Visual indicator (Closed/Open/Half-Open)
3. **Error Rate** - Percentage of failed requests
4. **Response Time** - P50, P95, P99 latencies
5. **Bulkhead Active Requests** - Current load vs capacity
6. **Bulkhead Rejected Requests** - Rate of rejections
7. **Order Status** - Success vs failure counts
8. **Inventory Levels** - Current stock levels
9. **Circuit Breaker Failures** - Failure rate per circuit
10. **Chaos Mode Status** - Which chaos modes are active

## 🎭 Demo Scenarios

Run the interactive demo:

```bash
./scripts/demo-scenarios.sh
```

### Available Scenarios

#### 1. Normal Operation
Tests the system under normal conditions with no failures.

```bash
curl -X POST http://localhost:8080/order/create \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"item_id": "item-1", "quantity": 1, "price": 999.99}
    ]
  }'
```

#### 2. Fail Fast Demo
Demonstrates input validation and immediate failure response.

```bash
# Invalid: Empty order
curl -X POST http://localhost:8080/order/create \
  -H "Content-Type: application/json" \
  -d '{"items": []}'

# Invalid: Negative quantity
curl -X POST http://localhost:8080/order/create \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"item_id": "item-1", "quantity": -1, "price": 999.99}
    ]
  }'
```

#### 3. Circuit Breaker Demo (Inventory)
Triggers circuit breaker by enabling inventory failures.

```bash
# Enable chaos
curl -X POST http://localhost:8081/chaos/inventory/enable

# Make multiple requests to trigger circuit breaker
for i in {1..10}; do
  curl -X POST http://localhost:8080/order/create \
    -H "Content-Type: application/json" \
    -d '{"items": [{"item_id": "item-1", "quantity": 1, "price": 999.99}]}'
  sleep 2
done

# Check circuit status
curl http://localhost:8080/order/circuit-status
```

**Expected behavior**:
- Initial requests fail due to inventory service errors
- After 60% failure rate, circuit opens
- Subsequent requests fail immediately (circuit open)
- After 30 seconds, circuit enters half-open state
- If service recovers, circuit closes

#### 4. Circuit Breaker Demo (Payment)
Similar to inventory but with payment service failures.

```bash
# Enable payment chaos
curl -X POST http://localhost:8082/chaos/payment/enable
curl -X POST http://localhost:8082/chaos/payment/slow
```

#### 5. Bulkhead Demo
Demonstrates resource isolation by sending concurrent requests.

```bash
# Enable slow mode
curl -X POST http://localhost:8082/chaos/payment/slow

# Send 15 concurrent requests (limit is 10)
for i in {1..15}; do
  curl -X POST http://localhost:8080/order/create \
    -H "Content-Type: application/json" \
    -d '{"items": [{"item_id": "item-1", "quantity": 1, "price": 999.99}]}' &
done
wait
```

**Expected behavior**:
- First 10 requests execute (may be slow due to payment delays)
- Remaining 5 requests are rejected by bulkhead
- System remains responsive to other operations

#### 6. Combined Chaos
Enables all failure modes simultaneously.

```bash
# Enable all chaos modes
curl -X POST http://localhost:8081/chaos/inventory/enable
curl -X POST http://localhost:8081/chaos/inventory/slow
curl -X POST http://localhost:8082/chaos/payment/enable
curl -X POST http://localhost:8082/chaos/payment/slow
```

#### 7. Reset All
Disables all chaos modes to restore normal operation.

```bash
# Disable all chaos modes (failures + slow mode)
curl -X POST http://localhost:8081/chaos/inventory/disable
curl -X POST http://localhost:8082/chaos/payment/disable

# Or disable slow mode only
curl -X POST http://localhost:8081/chaos/inventory/slow/disable
curl -X POST http://localhost:8082/chaos/payment/slow/disable
```

## 📡 API Reference

### Order Service

#### Create Order
```bash
POST /order/create
Content-Type: application/json

{
  "items": [
    {
      "item_id": "item-1",
      "quantity": 2,
      "price": 999.99
    }
  ]
}
```

#### Get Order
```bash
GET /order/:orderId
```

#### Circuit Breaker Status
```bash
GET /order/circuit-status
```

### Inventory Service

#### Check Inventory
```bash
GET /inventory/check/:itemId
```

#### Reserve Items
```bash
POST /inventory/reserve
Content-Type: application/json

{
  "order_id": "uuid",
  "items": [
    {
      "item_id": "item-1",
      "quantity": 2,
      "price": 999.99
    }
  ]
}
```

#### Chaos Engineering
```bash
POST /chaos/inventory/enable        # Enable 30% failure rate
POST /chaos/inventory/disable       # Disable all chaos modes
POST /chaos/inventory/slow          # Enable 2-5 second delays
POST /chaos/inventory/slow/disable  # Disable slow mode only
```

#### Health Check
```bash
GET /inventory/status
```

### Payment Service

#### Charge Payment
```bash
POST /payment/charge
Content-Type: application/json

{
  "order_id": "uuid",
  "amount": 1059.97
}
```

#### Chaos Engineering
```bash
POST /chaos/payment/enable        # Enable 40% failure rate
POST /chaos/payment/disable       # Disable all chaos modes
POST /chaos/payment/slow          # Enable 5-10 second delays
POST /chaos/payment/slow/disable  # Disable slow mode only
```

#### Health Check
```bash
GET /payment/status
```

### Metrics (All Services)
```bash
GET /metrics  # Prometheus format
```

## 📈 Metrics Collected

### HTTP Metrics
- `http_requests_total` - Total requests by service, method, endpoint, status
- `http_request_duration_seconds` - Request duration histogram

### Circuit Breaker Metrics
- `circuit_breaker_state` - Current state (0=closed, 1=open, 2=half-open)
- `circuit_breaker_failures_total` - Total failures per circuit

### Bulkhead Metrics
- `bulkhead_active_requests` - Current active requests
- `bulkhead_rejected_requests_total` - Total rejected requests

### Business Metrics
- `orders_total` - Total orders by status
- `inventory_level` - Current inventory per item
- `payment_amount_dollars` - Payment amount distribution

### Chaos Metrics
- `chaos_failure_enabled` - Whether chaos mode is active
- `chaos_slow_mode_enabled` - Whether slow mode is active

## 🏗️ Project Structure

```
resilience-demo/
├── services/
│   ├── order-service/
│   │   ├── main.go              # Order service with all patterns
│   │   └── Dockerfile
│   ├── inventory-service/
│   │   ├── main.go              # Inventory with chaos engineering
│   │   └── Dockerfile
│   └── payment-service/
│       ├── main.go              # Payment with chaos engineering
│       └── Dockerfile
├── internal/
│   ├── models/                  # Shared data models
│   │   ├── order.go
│   │   ├── inventory.go
│   │   └── payment.go
│   ├── patterns/                # Resilience pattern implementations
│   │   ├── circuitbreaker.go
│   │   ├── bulkhead.go
│   │   └── timeout.go
│   └── metrics/                 # Prometheus metrics
│       └── prometheus.go
├── monitoring/
│   ├── prometheus.yml           # Prometheus configuration
│   └── grafana/
│       ├── dashboards/          # Pre-configured dashboards
│       │   ├── dashboard.yml
│       │   └── resilience-patterns.json
│       └── datasources/         # Prometheus datasource
│           └── prometheus.yml
├── scripts/
│   ├── start-all.sh            # Start all services
│   ├── stop-all.sh             # Stop all services
│   └── demo-scenarios.sh       # Interactive demo
├── docker-compose.yml          # Service orchestration
├── go.mod                      # Go dependencies
├── go.sum
└── README.md
```

## 🔧 Configuration

### Circuit Breaker Settings
```go
MaxRequests: 3                      // Half-open state
Interval:    10 * time.Second       // Failure tracking window
Timeout:     30 * time.Second       // Time before half-open
ReadyToTrip: 60% failure rate with ≥3 requests
```

### Bulkhead Settings
```go
Size:    10                         // Max concurrent requests
Timeout: 1 * time.Second           // Acquire timeout
```

### HTTP Client Settings
```go
Timeout:     3 * time.Second       // Default timeout
RetryCount:  0                     // No automatic retries
```

### Chaos Engineering Settings

**Inventory Service:**
- Failure rate: 30%
- Slow mode delay: 2-5 seconds

**Payment Service:**
- Failure rate: 40%
- Slow mode delay: 5-10 seconds

## 🧪 Testing the Patterns

### Test Fail Fast
```bash
# Should fail immediately with validation error
time curl -X POST http://localhost:8080/order/create \
  -H "Content-Type: application/json" \
  -d '{"items": []}'

# Response should be < 100ms
```

### Test Circuit Breaker
```bash
# Enable failures
curl -X POST http://localhost:8081/chaos/inventory/enable

# Trigger circuit breaker (requires multiple failures)
for i in {1..10}; do
  curl -X POST http://localhost:8080/order/create \
    -H "Content-Type: application/json" \
    -d '{"items": [{"item_id": "item-1", "quantity": 1, "price": 999.99}]}'
  
  # Check circuit status
  curl http://localhost:8080/order/circuit-status | jq '.inventory_circuit.state'
  sleep 1
done

# After circuit opens, requests fail immediately (fast)
```

### Test Bulkhead
```bash
# Enable slow mode
curl -X POST http://localhost:8082/chaos/payment/slow

# Monitor active requests in Grafana while sending concurrent requests
for i in {1..20}; do
  curl -X POST http://localhost:8080/order/create \
    -H "Content-Type: application/json" \
    -d '{"items": [{"item_id": "item-1", "quantity": 1, "price": 999.99}]}' &
done
```

## 🔍 Observing the Patterns

### In Grafana (http://localhost:3000)

1. **Circuit Breaker Visualization**
   - Watch "Circuit Breaker State" gauge
   - Green (0) = Closed (healthy)
   - Red (1) = Open (failing)
   - Yellow (2) = Half-Open (testing recovery)

2. **Bulkhead Visualization**
   - "Bulkhead Active Requests" should max at 10
   - "Bulkhead Rejected Requests" increases when limit exceeded

3. **Response Time Impact**
   - Compare response times with/without chaos
   - Circuit open = fast failures
   - Circuit closed with chaos = slow failures

4. **Success Rate**
   - "Order Status" shows completed vs failed
   - With patterns: more predictable behavior
   - Without patterns: cascading failures

### In Prometheus (http://localhost:9090)

Example queries:
```promql
# Request rate
rate(http_requests_total[1m])

# Error rate
rate(http_requests_total{status=~"5.."}[1m]) / rate(http_requests_total[1m])

# Circuit breaker state
circuit_breaker_state

# Bulkhead saturation
bulkhead_active_requests / 10 * 100
```

## 🛠️ Development

### Running Locally (without Docker)

```bash
# Terminal 1: Inventory Service
cd services/inventory-service
go run main.go

# Terminal 2: Payment Service
cd services/payment-service
go run main.go

# Terminal 3: Order Service
cd services/order-service
export INVENTORY_SERVICE_URL=http://localhost:8081
export PAYMENT_SERVICE_URL=http://localhost:8082
go run main.go
```

### Running Prometheus & Grafana
```bash
# Just monitoring services
docker-compose up prometheus grafana
```

### Building Locally
```bash
# Download dependencies
go mod download

# Run tests (if implemented)
go test ./...

# Build all services
go build -o build/order-service ./services/order-service
go build -o build/inventory-service ./services/inventory-service
go build -o build/payment-service ./services/payment-service
```

## 📝 Key Learnings

### Without Resilience Patterns
- ❌ One slow service blocks entire system
- ❌ Cascading failures propagate
- ❌ No resource isolation
- ❌ Poor user experience during failures

### With Resilience Patterns
- ✅ **Fail Fast**: Quick feedback on invalid requests
- ✅ **Circuit Breaker**: Prevents cascade, allows recovery
- ✅ **Bulkhead**: Isolates failures, protects resources
- ✅ Better observability with metrics
- ✅ Graceful degradation

## 📚 Dependencies

- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/sony/gobreaker` - Circuit breaker implementation
- `github.com/go-resty/resty/v2` - HTTP client
- `github.com/prometheus/client_golang` - Prometheus metrics
- `github.com/sirupsen/logrus` - Structured logging

## 🧹 Cleanup

```bash
# Stop all services
./scripts/stop-all.sh

# Remove volumes
docker-compose down -v

# Remove images
docker-compose down --rmi all
```

## 🐛 Troubleshooting

### Services not starting
```bash
# Check Docker daemon
docker info

# View logs
docker-compose logs -f [service-name]

# Restart services
docker-compose restart
```

### Grafana dashboard not showing data
```bash
# Check Prometheus targets
open http://localhost:9090/targets

# All should be "UP"
# If not, check service logs
```

### Circuit breaker not opening
- Ensure chaos mode is enabled
- Need at least 3 requests with 60% failure rate
- Check circuit status: `curl http://localhost:8080/order/circuit-status`

### Port conflicts
```bash
# Check what's using ports
lsof -i :8080
lsof -i :8081
lsof -i :8082
lsof -i :9090
lsof -i :3000
```

## 📄 License

MIT License - Feel free to use for learning and demonstration purposes.

## 🤝 Contributing

This is a demonstration project for educational purposes. Feel free to:
- Add more resilience patterns (Retry, Rate Limiting, etc.)
- Improve metrics and dashboards
- Add more chaos scenarios
- Implement automated tests

## 📖 Further Reading

- [Release It! by Michael Nygard](https://pragprog.com/titles/mnee2/release-it-second-edition/)
- [Circuit Breaker Pattern - Martin Fowler](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Bulkhead Pattern - Microsoft](https://docs.microsoft.com/en-us/azure/architecture/patterns/bulkhead)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)

---

**Happy Testing! 🚀**

For questions or issues, check the logs or create an issue in the repository.

