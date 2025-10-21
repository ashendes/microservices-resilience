output "namespace_id" {
  description = "Service discovery namespace ID"
  value       = aws_service_discovery_private_dns_namespace.this.id
}

output "namespace_name" {
  description = "Service discovery namespace name"
  value       = aws_service_discovery_private_dns_namespace.this.name
}

output "inventory_service_arn" {
  description = "Service discovery ARN for inventory service"
  value       = aws_service_discovery_service.inventory.arn
}

output "payment_service_arn" {
  description = "Service discovery ARN for payment service"
  value       = aws_service_discovery_service.payment.arn
}

output "order_service_arn" {
  description = "Service discovery ARN for order service"
  value       = aws_service_discovery_service.order.arn
}

