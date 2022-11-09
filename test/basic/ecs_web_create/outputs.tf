output "vpc_id" {
  value = aptible_aws_vpc.vpc.id
}

output "ecs_name" {
  value = aptible_aws_ecs_web.web.name
}

output "ecs_web_id" {
  value = aptible_aws_ecs_web.web.id
}

output "certificate_arn" {
  value = aptible_aws_acm.cert.arn
}

output "web_url" {
  value = aptible_aws_acm.cert.fqdn
}
output "loadbalancer_url" {
  value = aptible_aws_ecs_web.web.load_balancer_url
}

output "aptible_aws_account_id" {
  value = data.aptible_environment.env.aws_account_id
}
