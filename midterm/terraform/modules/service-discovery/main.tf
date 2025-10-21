# Create a private DNS namespace for service discovery
resource "aws_service_discovery_private_dns_namespace" "this" {
  name        = var.namespace_name
  description = "Private DNS namespace for microservices"
  vpc         = var.vpc_id
}

# Create service discovery service for inventory
resource "aws_service_discovery_service" "inventory" {
  name = "inventory"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.this.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {}
}

# Create service discovery service for payment
resource "aws_service_discovery_service" "payment" {
  name = "payment"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.this.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {}
}

# Create service discovery service for order
resource "aws_service_discovery_service" "order" {
  name = "order"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.this.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {}
}

