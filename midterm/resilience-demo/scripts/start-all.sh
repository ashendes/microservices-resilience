#!/bin/bash

echo "🚀 Starting Resilience Patterns Demo..."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running. Please start Docker first."
    exit 1
fi

# Build and start services
echo "📦 Building and starting services..."
docker-compose up --build -d

# Wait for services to be ready
echo ""
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check service health
echo ""
echo "🔍 Checking service health..."

services=("order-service:8080" "inventory-service:8081" "payment-service:8082" "prometheus:9090" "grafana:3000")
all_healthy=true

for service in "${services[@]}"; do
    IFS=':' read -r name port <<< "$service"
    if curl -s -o /dev/null -w "%{http_code}" "http://localhost:$port" | grep -q "200\|302"; then
        echo "✅ $name is healthy (port $port)"
    else
        echo "⚠️  $name may not be ready yet (port $port)"
        all_healthy=false
    fi
done

echo ""
echo "🎉 Resilience Patterns Demo is running!"
echo ""
echo "📊 Access points:"
echo "  - Order Service:     http://localhost:8080"
echo "  - Inventory Service: http://localhost:8081"
echo "  - Payment Service:   http://localhost:8082"
echo "  - Prometheus:        http://localhost:9090"
echo "  - Grafana:           http://localhost:3000 (admin/admin)"
echo ""
echo "📖 View logs with: docker-compose logs -f [service-name]"
echo "🛑 Stop all services: ./scripts/stop-all.sh"
echo "🎭 Run demo scenarios: ./scripts/demo-scenarios.sh"
echo ""

