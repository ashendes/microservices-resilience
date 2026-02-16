FROM grafana/grafana:latest

USER root

# Copy dashboard provider configuration
COPY dashboards/dashboard.yml /etc/grafana/provisioning/dashboards/

# Copy dashboard JSON files
COPY dashboards/*.json /etc/grafana/provisioning/dashboards/

# Copy custom entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

USER grafana

# Use custom entrypoint that configures datasource at runtime
ENTRYPOINT ["/entrypoint.sh"]

