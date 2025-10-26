#!/bin/bash

BASE_URL="http://localhost"
ORDER_SERVICE="$BASE_URL:8080"
INVENTORY_SERVICE="$BASE_URL:8081"
PAYMENT_SERVICE="$BASE_URL:8082"

echo "🎭 Resilience Patterns Demo Scenarios"
echo "======================================"
echo ""

# Function to create an order
create_order() {
    echo "📦 Creating order..."
    response=$(curl -s -X POST "$ORDER_SERVICE/order/create" \
        -H "Content-Type: application/json" \
        -d '{
            "items": [
                {"item_id": "item-1", "quantity": 1, "price": 999.99},
                {"item_id": "item-2", "quantity": 2, "price": 29.99}
            ]
        }')
    echo "$response" | jq '.'
    echo ""
}

# Function to check circuit breaker status
check_circuits() {
    echo "🔌 Circuit Breaker Status:"
    curl -s "$ORDER_SERVICE/order/circuit-status" | jq '.'
    echo ""
}

# Function to check service health
check_health() {
    echo "🏥 Service Health:"
    echo "  Inventory: $(curl -s "$INVENTORY_SERVICE/inventory/status" | jq -r '.status')"
    echo "  Payment:   $(curl -s "$PAYMENT_SERVICE/payment/status" | jq -r '.status')"
    echo ""
}

echo "Select a demo scenario:"
echo "1. Normal Operation (No Failures)"
echo "2. Fail Fast Demo (Invalid Orders)"
echo "3. Circuit Breaker Demo (Inventory Failures)"
echo "4. Circuit Breaker Demo (Payment Failures)"
echo "5. Bulkhead Demo (Concurrent Requests)"
echo "6. Combined Chaos (Multiple Failures)"
echo "7. Reset All (Disable Chaos)"
echo "8. Check Status (Circuits & Health)"
echo ""
read -p "Enter scenario number (1-8): " scenario

case $scenario in
    1)
        echo ""
        echo "=== Scenario 1: Normal Operation ==="
        echo ""
        echo "Disabling all chaos modes..."
        curl -s -X POST "$INVENTORY_SERVICE/chaos/inventory/disable" > /dev/null
        curl -s -X POST "$PAYMENT_SERVICE/chaos/payment/disable" > /dev/null
        echo "✅ Chaos disabled"
        echo ""
        
        echo "Creating 5 successful orders..."
        for i in {1..10}; do
            echo "Order $i:"
            create_order
            # sleep 1
        done
        
        check_circuits
        ;;
    
    2)
        echo ""
        echo "=== Scenario 2: Fail Fast Demo ==="
        echo ""
        echo "Testing input validation (Fail Fast pattern)..."
        echo ""
        
        echo "❌ Test 1: Empty order"
        curl -s -X POST "$ORDER_SERVICE/order/create" \
            -H "Content-Type: application/json" \
            -d '{"items": []}' | jq '.'
        echo ""
        
        echo "❌ Test 2: Invalid quantity"
        curl -s -X POST "$ORDER_SERVICE/order/create" \
            -H "Content-Type: application/json" \
            -d '{
                "items": [
                    {"item_id": "item-1", "quantity": 0, "price": 999.99}
                ]
            }' | jq '.'
        echo ""
        
        echo "❌ Test 3: Invalid price"
        curl -s -X POST "$ORDER_SERVICE/order/create" \
            -H "Content-Type: application/json" \
            -d '{
                "items": [
                    {"item_id": "item-1", "quantity": 1, "price": -10}
                ]
            }' | jq '.'
        echo ""
        
        echo "✅ All requests failed fast without calling downstream services!"
        ;;
    
    3)
        echo ""
        echo "=== Scenario 3: Circuit Breaker Demo (Inventory) ==="
        echo ""
        
        echo "Enabling inventory chaos (30% failure rate)..."
        curl -s -X POST "$INVENTORY_SERVICE/chaos/inventory/enable" > /dev/null
        echo "✅ Chaos enabled"
        echo ""
        
        echo "Sending requests to trigger circuit breaker..."
        for i in {1..40}; do
            echo "Request $i:"
            create_order
            # check_circuits
            # sleep 2
        done
        
        echo "🔴 Notice: Circuit breaker should open after repeated failures"
        echo "🟡 Then: Circuit breaker enters half-open state after timeout"
        echo "🟢 Finally: Circuit breaker closes if service recovers"
        ;;
    
    4)
        echo ""
        echo "=== Scenario 4: Circuit Breaker Demo (Payment) ==="
        echo ""
        
        echo "Enabling payment chaos (40% failure rate + slow responses)..."
        curl -s -X POST "$PAYMENT_SERVICE/chaos/payment/enable" > /dev/null
        curl -s -X POST "$PAYMENT_SERVICE/chaos/payment/slow" > /dev/null
        echo "✅ Chaos enabled"
        echo ""
        
        echo "Sending requests to trigger circuit breaker..."
        for i in {1..40}; do
            echo "Request $i:"
            create_order
            # check_circuits
            # sleep 2
        done
        
        echo "💡 Watch Grafana dashboard to see circuit breaker state changes"
        ;;
    
    5)
        echo ""
        echo "=== Scenario 5: Bulkhead Demo ==="
        echo ""
        
        echo "Enabling slow mode on payment service..."
        curl -s -X POST "$PAYMENT_SERVICE/chaos/payment/slow" > /dev/null
        echo "✅ Slow mode enabled (5-10 second delays)"
        echo ""
        
        echo "Sending 15 concurrent requests (bulkhead limit is 10)..."
        echo "Expected: 10 requests in progress, 5 rejected by bulkhead"
        echo ""
        
        for i in {1..15}; do
            create_order &
        done
        
        wait
        echo ""
        echo "💡 Check Grafana to see bulkhead metrics:"
        echo "   - Active requests should max at 10"
        echo "   - Rejected requests counter should increment"
        ;;
    
    6)
        echo ""
        echo "=== Scenario 6: Combined Chaos ==="
        echo ""
        
        echo "Enabling all chaos modes..."
        curl -s -X POST "$INVENTORY_SERVICE/chaos/inventory/enable" > /dev/null
        curl -s -X POST "$INVENTORY_SERVICE/chaos/inventory/slow" > /dev/null
        curl -s -X POST "$PAYMENT_SERVICE/chaos/payment/enable" > /dev/null
        curl -s -X POST "$PAYMENT_SERVICE/chaos/payment/slow" > /dev/null
        echo "✅ All chaos enabled"
        echo ""
        
        check_health
        
        echo "Sending requests with all failures enabled..."
        for i in {1..20}; do
            echo "Request $i:"
            create_order
            if [ $((i % 5)) -eq 0 ]; then
                check_circuits
            fi
            sleep 1
        done
        
        echo "💥 Observe how resilience patterns handle multiple failure modes:"
        echo "   - Fail Fast: Quick validation before calling services"
        echo "   - Circuit Breaker: Prevents cascading failures"
        echo "   - Bulkhead: Limits resource consumption"
        ;;
    
    7)
        echo ""
        echo "=== Scenario 7: Reset All ==="
        echo ""
        
        echo "Disabling all chaos modes..."
        curl -s -X POST "$INVENTORY_SERVICE/chaos/inventory/disable" > /dev/null
        curl -s -X POST "$PAYMENT_SERVICE/chaos/payment/disable" > /dev/null
        echo "✅ All chaos disabled"
        echo ""
        
        check_health
        check_circuits
        
        echo "System reset to normal operation"
        ;;
    
    8)
        echo ""
        echo "=== Scenario 8: Status Check ==="
        echo ""
        
        check_health
        check_circuits
        
        echo "📊 View detailed metrics at:"
        echo "   - Prometheus: http://localhost:9090"
        echo "   - Grafana:    http://localhost:3000"
        ;;
    
    *)
        echo "Invalid scenario number"
        exit 1
        ;;
esac

echo ""
echo "✅ Scenario complete!"
echo ""
echo "💡 Tips:"
echo "   - Watch Grafana dashboard for real-time metrics"
echo "   - View logs: docker-compose logs -f [service-name]"
echo "   - Run another scenario: ./scripts/demo-scenarios.sh"
echo ""

