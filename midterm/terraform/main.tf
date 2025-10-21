# Wire together modules: network, ecr, logging, ecs, service-discovery

# Network configuration (shared by all services)
module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = 8080  # Base port, actual services use 8080-8082
}

# Service Discovery disabled due to AWS Academy permission restrictions
# Using internal ALBs instead for service-to-service communication

# Shared ECS Cluster for all microservices
resource "aws_ecs_cluster" "shared" {
  name = "${var.service_name}-cluster"
}

# Reuse an existing IAM role for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# Single ALB for all services
resource "aws_lb" "shared" {
  name               = "${var.service_name}-alb"
  load_balancer_type = "application"
  subnets            = module.network.subnet_ids
  security_groups    = [module.network.security_group_id]
}

# Target group for order service (port 8080)
resource "aws_lb_target_group" "order" {
  name        = "${var.service_name}-order-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = module.network.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 30
    matcher             = "200"
    timeout             = 5
    path                = "/health"
  }
}

# Target group for inventory service (port 8081)
resource "aws_lb_target_group" "inventory" {
  name        = "${var.service_name}-inventory-tg"
  port        = 8081
  protocol    = "HTTP"
  vpc_id      = module.network.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 30
    matcher             = "200"
    timeout             = 5
    path                = "/health"
  }
}

# Target group for payment service (port 8082)
resource "aws_lb_target_group" "payment" {
  name        = "${var.service_name}-payment-tg"
  port        = 8082
  protocol    = "HTTP"
  vpc_id      = module.network.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 30
    matcher             = "200"
    timeout             = 5
    path                = "/health"
  }
}

# Target group for Grafana (port 3000)
resource "aws_lb_target_group" "grafana" {
  name        = "${var.service_name}-grafana-tg"
  port        = 3000
  protocol    = "HTTP"
  vpc_id      = module.network.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 30
    matcher             = "200"
    timeout             = 5
    path                = "/api/health"
  }
}

# Target group for Prometheus (port 9090)
resource "aws_lb_target_group" "prometheus" {
  name        = "${var.service_name}-prom-tg"
  port        = 9090
  protocol    = "HTTP"
  vpc_id      = module.network.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 30
    matcher             = "200,301,302"
    timeout             = 5
    path                = "/-/healthy"
  }
}

# Listener for order service (port 80 -> 8080)
resource "aws_lb_listener" "order" {
  load_balancer_arn = aws_lb.shared.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.order.arn
  }
}

# Listener for inventory service (port 8081 -> 8081)
resource "aws_lb_listener" "inventory" {
  load_balancer_arn = aws_lb.shared.arn
  port              = "8081"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.inventory.arn
  }
}

# Listener for payment service (port 8082 -> 8082)
resource "aws_lb_listener" "payment" {
  load_balancer_arn = aws_lb.shared.arn
  port              = "8082"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.payment.arn
  }
}

# Listener for Grafana (port 3000)
resource "aws_lb_listener" "grafana" {
  load_balancer_arn = aws_lb.shared.arn
  port              = "3000"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.grafana.arn
  }
}

# Listener for Prometheus (port 9090)
resource "aws_lb_listener" "prometheus" {
  load_balancer_arn = aws_lb.shared.arn
  port              = "9090"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.prometheus.arn
  }
}

# ========================================
# INVENTORY SERVICE
# ========================================

module "ecr_inventory" {
  source          = "./modules/ecr"
  repository_name = "${var.service_name}-inventory"
}

module "logging_inventory" {
  source            = "./modules/logging"
  service_name      = "${var.service_name}-inventory"
  retention_in_days = var.log_retention_days
}

module "ecs_inventory" {
  source             = "./modules/ecs"
  service_name       = "${var.service_name}-inventory"
  cluster_id         = aws_ecs_cluster.shared.id
  cluster_name       = aws_ecs_cluster.shared.name
  image              = "${module.ecr_inventory.repository_url}:latest"
  container_port     = 8081
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging_inventory.log_group_name
  ecs_count          = var.ecs_count
  cpu                = var.cpu
  memory             = var.memory
  region             = var.aws_region
  target_group_arn   = aws_lb_target_group.inventory.arn
  environment_variables = []
  
  depends_on = [
    aws_lb_listener.inventory,
    docker_registry_image.inventory
  ]
}

# Build & push inventory service image
resource "docker_image" "inventory" {
  name = "${module.ecr_inventory.repository_url}:latest"
  build {
    context    = "../resilience-demo"
    dockerfile = "services/inventory-service/Dockerfile"
  }
}

resource "docker_registry_image" "inventory" {
  name = docker_image.inventory.name
}

# ========================================
# PAYMENT SERVICE
# ========================================

module "ecr_payment" {
  source          = "./modules/ecr"
  repository_name = "${var.service_name}-payment"
}

module "logging_payment" {
  source            = "./modules/logging"
  service_name      = "${var.service_name}-payment"
  retention_in_days = var.log_retention_days
}

module "ecs_payment" {
  source             = "./modules/ecs"
  service_name       = "${var.service_name}-payment"
  cluster_id         = aws_ecs_cluster.shared.id
  cluster_name       = aws_ecs_cluster.shared.name
  image              = "${module.ecr_payment.repository_url}:latest"
  container_port     = 8082
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging_payment.log_group_name
  ecs_count          = var.ecs_count
  cpu                = var.cpu
  memory             = var.memory
  region             = var.aws_region
  target_group_arn   = aws_lb_target_group.payment.arn
  environment_variables = []
  
  depends_on = [
    aws_lb_listener.payment,
    docker_registry_image.payment
  ]
}

# Build & push payment service image
resource "docker_image" "payment" {
  name = "${module.ecr_payment.repository_url}:latest"
  build {
    context    = "../resilience-demo"
    dockerfile = "services/payment-service/Dockerfile"
  }
}

resource "docker_registry_image" "payment" {
  name = docker_image.payment.name
}

# ========================================
# ORDER SERVICE
# ========================================

module "ecr_order" {
  source          = "./modules/ecr"
  repository_name = "${var.service_name}-order"
}

module "logging_order" {
  source            = "./modules/logging"
  service_name      = "${var.service_name}-order"
  retention_in_days = var.log_retention_days
}

module "ecs_order" {
  source             = "./modules/ecs"
  service_name       = "${var.service_name}-order"
  cluster_id         = aws_ecs_cluster.shared.id
  cluster_name       = aws_ecs_cluster.shared.name
  image              = "${module.ecr_order.repository_url}:latest"
  container_port     = 8080
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging_order.log_group_name
  ecs_count          = var.ecs_count
  cpu                = var.cpu
  memory             = var.memory
  target_group_arn   = aws_lb_target_group.order.arn
  region             = var.aws_region
  environment_variables = [
    {
      name  = "INVENTORY_SERVICE_URL"
      value = "http://${aws_lb.shared.dns_name}:8081"
    },
    {
      name  = "PAYMENT_SERVICE_URL"
      value = "http://${aws_lb.shared.dns_name}:8082"
    }
  ]
  
  depends_on = [
    aws_lb_listener.order,
    aws_lb_listener.inventory,
    aws_lb_listener.payment,
    docker_registry_image.order,
    docker_registry_image.inventory,
    docker_registry_image.payment
  ]
}

# Build & push order service image
resource "docker_image" "order" {
  name = "${module.ecr_order.repository_url}:latest"
  build {
    context    = "../resilience-demo"
    dockerfile = "services/order-service/Dockerfile"
  }
}

resource "docker_registry_image" "order" {
  name = docker_image.order.name
}

# ========================================
# MONITORING STACK
# ========================================

# Prometheus configuration with ALB DNS substituted
# locals {
#   prometheus_config = templatefile("${path.module}/prometheus-config.yml", {
#     alb_dns_name = aws_lb.shared.dns_name
#   })
# }

# ECR repository for custom Grafana image
module "ecr_grafana" {
  source          = "./modules/ecr"
  repository_name = "${var.service_name}-grafana"
}

# CloudWatch log group for Prometheus
resource "aws_cloudwatch_log_group" "prometheus" {
  name              = "/ecs/${var.service_name}-prometheus"
  retention_in_days = var.log_retention_days
}

# CloudWatch log group for Grafana
resource "aws_cloudwatch_log_group" "grafana" {
  name              = "/ecs/${var.service_name}-grafana"
  retention_in_days = var.log_retention_days
}

# Prometheus ECS task definition
resource "aws_ecs_task_definition" "prometheus" {
  family                   = "${var.service_name}-prometheus-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = data.aws_iam_role.lab_role.arn
  task_role_arn            = data.aws_iam_role.lab_role.arn

  container_definitions = jsonencode([{
    name      = "${var.service_name}-prometheus"
    image     = "prom/prometheus:latest"
    essential = true

    command = [
      "--config.file=/etc/prometheus/prometheus.yml",
      "--storage.tsdb.path=/prometheus",
      "--web.console.libraries=/usr/share/prometheus/console_libraries",
      "--web.console.templates=/usr/share/prometheus/consoles"
    ]

    # environment = [
    #   {
    #     name  = "PROMETHEUS_CONFIG"
    #     value = local.prometheus_config
    #   }
    # ]

    portMappings = [{
      containerPort = 9090
      protocol      = "tcp"
    }]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.prometheus.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "prometheus"
      }
    }

    mountPoints = []
    volumesFrom = []
  }])
}

# Prometheus ECS service
resource "aws_ecs_service" "prometheus" {
  name            = "${var.service_name}-prometheus"
  cluster         = aws_ecs_cluster.shared.id
  task_definition = aws_ecs_task_definition.prometheus.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = module.network.subnet_ids
    security_groups  = [module.network.security_group_id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.prometheus.arn
    container_name   = "${var.service_name}-prometheus"
    container_port   = 9090
  }

  depends_on = [
    aws_lb_listener.prometheus,
    module.ecs_order,
    module.ecs_inventory,
    module.ecs_payment
  ]
}

# Build & push custom Grafana image with dashboards
resource "docker_image" "grafana" {
  name = "${module.ecr_grafana.repository_url}:latest"
  build {
    context    = "../resilience-demo/monitoring/grafana"
    dockerfile = "Dockerfile"
    platform   = "linux/amd64"
  }
}

resource "docker_registry_image" "grafana" {
  name = docker_image.grafana.name
}

# Grafana ECS task definition
resource "aws_ecs_task_definition" "grafana" {
  family                   = "${var.service_name}-grafana-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = data.aws_iam_role.lab_role.arn
  task_role_arn            = data.aws_iam_role.lab_role.arn

  container_definitions = jsonencode([{
    name      = "${var.service_name}-grafana"
    image     = "${module.ecr_grafana.repository_url}:latest"
    essential = true

    environment = [
      {
        name  = "GF_SECURITY_ADMIN_PASSWORD"
        value = "admin"
      },
      {
        name  = "GF_SECURITY_ADMIN_USER"
        value = "admin"
      },
      {
        name  = "GF_USERS_ALLOW_SIGN_UP"
        value = "false"
      },
      {
        name  = "GF_AUTH_ANONYMOUS_ENABLED"
        value = "false"
      },
      {
        name  = "GF_AUTH_ANONYMOUS_ORG_ROLE"
        value = "Viewer"
      },
      {
        name  = "GF_INSTALL_PLUGINS"
        value = ""
      },
      {
        name  = "PROMETHEUS_URL"
        value = "http://${aws_lb.shared.dns_name}:9090"
      }
    ]

    portMappings = [{
      containerPort = 3000
      protocol      = "tcp"
    }]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.grafana.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "grafana"
      }
    }

    mountPoints = []
    volumesFrom = []
  }])
}

# Grafana ECS service
resource "aws_ecs_service" "grafana" {
  name            = "${var.service_name}-grafana"
  cluster         = aws_ecs_cluster.shared.id
  task_definition = aws_ecs_task_definition.grafana.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = module.network.subnet_ids
    security_groups  = [module.network.security_group_id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.grafana.arn
    container_name   = "${var.service_name}-grafana"
    container_port   = 3000
  }

  depends_on = [
    aws_lb_listener.grafana,
    docker_registry_image.grafana,
    aws_ecs_service.prometheus
  ]
}
