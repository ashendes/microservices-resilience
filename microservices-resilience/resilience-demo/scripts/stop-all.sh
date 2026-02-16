#!/bin/bash

echo "🛑 Stopping Resilience Patterns Demo..."
echo ""

docker-compose down

echo ""
echo "✅ All services stopped."
echo ""
echo "💡 To remove volumes as well, run: docker-compose down -v"
echo ""

