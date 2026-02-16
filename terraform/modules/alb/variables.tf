variable "lb_name" {
  type        = string
  description = "Base name for ALB resources"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnets for FARGATE tasks"
}

variable "security_group_ids" {
  type        = list(string)
  description = "SGs for FARGATE tasks"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID"
}

variable "internal" {
  type        = bool
  description = "Whether the ALB is internal or internet-facing"
  default     = false
}

variable "target_port" {
  type        = number
  description = "Port for the target group"
  default     = 80
}
