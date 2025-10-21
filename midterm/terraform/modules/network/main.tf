# Fetch the default VPC
data "aws_vpc" "default" {
  default = true
}

# List all subnets in that VPC
data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# Create a security group to allow HTTP to ALB and container ports
resource "aws_security_group" "this" {
  name        = "${var.service_name}-sg"
  description = "Allow inbound on port 80 and microservice ports"
  vpc_id      = data.aws_vpc.default.id

  # Allow HTTP traffic to ALB on port 80
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = var.cidr_blocks
    description = "Allow HTTP traffic to ALB"
  }

  # Allow traffic to all microservice ports via ALB (8080-8082)
  ingress {
    from_port   = 8080
    to_port     = 8082
    protocol    = "tcp"
    cidr_blocks = var.cidr_blocks
    description = "Allow traffic to all services via ALB"
  }

  # Allow traffic to Grafana (port 3000)
  ingress {
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    cidr_blocks = var.cidr_blocks
    description = "Allow traffic to Grafana"
  }

  # Allow traffic to Prometheus (port 9090)
  ingress {
    from_port   = 9090
    to_port     = 9090
    protocol    = "tcp"
    cidr_blocks = var.cidr_blocks
    description = "Allow traffic to Prometheus"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }
}
