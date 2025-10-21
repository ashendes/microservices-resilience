output "alb_dns_name" {
  description = "DNS name of the shared Application Load Balancer"
  value       = aws_lb.shared.dns_name
}

output "order_service_url" {
  description = "Public URL to access the order service"
  value       = "http://${aws_lb.shared.dns_name}"
}

output "inventory_service_url" {
  description = "URL to access inventory service (same ALB, port 8081)"
  value       = "http://${aws_lb.shared.dns_name}:8081"
}

output "payment_service_url" {
  description = "URL to access payment service (same ALB, port 8082)"
  value       = "http://${aws_lb.shared.dns_name}:8082"
}

output "grafana_url" {
  description = "URL to access Grafana dashboard"
  value       = "http://${aws_lb.shared.dns_name}:3000"
}

output "prometheus_url" {
  description = "URL to access Prometheus"
  value       = "http://${aws_lb.shared.dns_name}:9090"
}

output "grafana_credentials" {
  description = "Grafana login credentials"
  value = {
    username = "admin"
    password = "admin"
  }
  sensitive = false
}

output "ecs_cluster_name" {
  description = "Name of the shared ECS cluster"
  value       = aws_ecs_cluster.shared.name
}

output "ecs_cluster_id" {
  description = "ID of the shared ECS cluster"
  value       = aws_ecs_cluster.shared.id
}

output "service_names" {
  description = "Names of all ECS services in the cluster"
  value = {
    order     = module.ecs_order.service_name
    inventory = module.ecs_inventory.service_name
    payment   = module.ecs_payment.service_name
  }
}

output "ecr_repositories" {
  description = "ECR repository URLs"
  value = {
    order     = module.ecr_order.repository_url
    inventory = module.ecr_inventory.repository_url
    payment   = module.ecr_payment.repository_url
    grafana   = module.ecr_grafana.repository_url
  }
}

output "monitoring_stack" {
  description = "Monitoring stack information"
  value = {
    grafana_url    = "http://${aws_lb.shared.dns_name}:3000"
    prometheus_url = "http://${aws_lb.shared.dns_name}:9090"
    username       = "admin"
    password       = "admin"
  }
}