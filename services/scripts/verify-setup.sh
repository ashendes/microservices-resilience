#!/bin/bash

echo "🔍 Verifying Resilience Demo Setup"
echo "===================================="
echo ""

errors=0

# Check if Docker is installed
echo "Checking Docker..."
if command -v docker &> /dev/null; then
    echo "✅ Docker is installed: $(docker --version)"
else
    echo "❌ Docker is not installed"
    errors=$((errors + 1))
fi

# Check if Docker Compose is installed
echo "Checking Docker Compose..."
if command -v docker-compose &> /dev/null; then
    echo "✅ Docker Compose is installed: $(docker-compose --version)"
else
    echo "❌ Docker Compose is not installed"
    errors=$((errors + 1))
fi

# Check if Go is installed
echo "Checking Go..."
if command -v go &> /dev/null; then
    echo "✅ Go is installed: $(go version)"
else
    echo "⚠️  Go is not installed (optional for Docker-only usage)"
fi

# Check if jq is installed
echo "Checking jq..."
if command -v jq &> /dev/null; then
    echo "✅ jq is installed: $(jq --version)"
else
    echo "⚠️  jq is not installed (optional, needed for demo scripts)"
fi

# Check if curl is installed
echo "Checking curl..."
if command -v curl &> /dev/null; then
    echo "✅ curl is installed: $(curl --version | head -n 1)"
else
    echo "❌ curl is not installed"
    errors=$((errors + 1))
fi

echo ""
echo "Checking project structure..."

# Check critical files
files=(
    "go.mod"
    "go.sum"
    "docker-compose.yml"
    "README.md"
    "services/order-service/main.go"
    "services/inventory-service/main.go"
    "services/payment-service/main.go"
    "internal/models/order.go"
    "internal/patterns/circuitbreaker.go"
    "internal/metrics/prometheus.go"
    "monitoring/prometheus.yml"
    "monitoring/grafana/dashboards/resilience-patterns.json"
    "scripts/start-all.sh"
    "scripts/stop-all.sh"
    "scripts/demo-scenarios.sh"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file"
    else
        echo "❌ Missing: $file"
        errors=$((errors + 1))
    fi
done

echo ""
echo "Checking script permissions..."
if [ -x "scripts/start-all.sh" ]; then
    echo "✅ scripts/start-all.sh is executable"
else
    echo "⚠️  scripts/start-all.sh is not executable (run: chmod +x scripts/*.sh)"
fi

echo ""
echo "===================================="
if [ $errors -eq 0 ]; then
    echo "✅ All checks passed! Setup is complete."
    echo ""
    echo "Next steps:"
    echo "  1. Run: ./scripts/start-all.sh"
    echo "  2. Open: http://localhost:3000 (Grafana)"
    echo "  3. Run: ./scripts/demo-scenarios.sh"
else
    echo "❌ Found $errors error(s). Please fix them before proceeding."
    exit 1
fi
echo ""

