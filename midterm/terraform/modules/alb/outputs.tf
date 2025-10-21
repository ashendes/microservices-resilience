output "lb_arn" {
  description = "ALB ARN"
  value       = aws_lb.this.arn
}

output "lb_dns_name" {
  description = "DNS name of the ALB"
  value       = aws_lb.this.dns_name
}

output "target_group_arn" {
  description = "ALB target group ARN"
  value       = aws_lb_target_group.this.arn
}
