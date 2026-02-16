output "cluster_name" {
  description = "ECS cluster name"
  value       = var.cluster_name
}

output "cluster_id" {
  description = "ECS cluster ID"
  value       = var.cluster_id
}

output "service_name" {
  description = "ECS service name"
  value       = aws_ecs_service.this.name
}
