# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-east-1"
}

# Base service name (will be prefixed to all services)
variable "service_name" {
  type    = string
  default = "microservices-resilience"
}

# ECS settings for all services
variable "ecs_count" {
  type        = number
  default     = 1
  description = "Number of tasks per service"
}

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}

# Existing IAM role name for ECS tasks
variable "ecs_task_role_name" {
  type        = string
  default     = "ecsTaskExecutionRole"
  description = "Name of an existing IAM role for ECS tasks"
}

# Resource allocation per service
variable "cpu" {
  type        = string
  default     = "512"
  description = "vCPU units (256, 512, 1024, 2048)"
}

variable "memory" {
  type        = string
  default     = "1024"
  description = "Memory in MiB (512, 1024, 2048, 3072, 4096)"
}

# Grafana admin credentials (set via tfvars or environment)
variable "grafana_admin_user" {
  type        = string
  default     = "change-me"
  description = "Grafana admin username"
}

variable "grafana_admin_password" {
  type        = string
  default     = "change-me"
  description = "Grafana admin password"
}