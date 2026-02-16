#!/bin/bash
set -e

# Create datasources directory if it doesn't exist
mkdir -p /etc/grafana/provisioning/datasources

# Create datasource configuration from environment variable
cat > /etc/grafana/provisioning/datasources/prometheus.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: ${PROMETHEUS_URL}
    isDefault: true
    editable: true
EOF

# Run the original Grafana entrypoint
exec /run.sh "$@"

