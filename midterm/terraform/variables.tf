# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-east-1"
}

# Base service name (will be prefixed to all services)
variable "service_name" {
  type    = string
  default = "hw7-resilience"
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