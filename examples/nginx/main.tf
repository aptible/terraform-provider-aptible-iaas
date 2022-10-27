terraform {
  required_providers {
    aptible = {
      source  = "aptible.com/aptible/aptible-iaas"
    }
    aws = {
      source = "hashicorp/aws"
    }
  }
}

variable "organization_id" {
  type    = string
}

variable "environment_id" {
  type    = string
}

variable "aptible_host" {
  type    = string
}

variable "fqdn" {
  type    = string
}

variable "domain" {
  type = string
}

provider "aws" {
  region = "us-east-1"
}

provider "aptible" {
  host = var.aptible_host
  token = var.aptible_token
}

data "aptible_aws_vpc" "network" {
  environment_id  = var.environment_id
  organization_id = var.organization_id
  name = ""
}

resource "aptible_aws_acm" "cert" {
  environment_id  = var.environment_id
  organization_id = var.organization_id
  fqdn              = var.fqdn
  validation_method = "DNS" # optional
}

data "aws_route53_zone" "domains" {
  name         = var.domain
  private_zone = false
}

locals {
  validation_dns = [
    for dvo in aptible_aws_acm.cert.domain_validation_records : {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  ]
}

resource "aws_route53_record" "domains" {
  allow_overwrite = true
  name            = local.validation_dns.0.name
  records         = [local.validation_dns.0.record]
  ttl             = 60
  type            = local.validation_dns.0.type
  zone_id         = data.aws_route53_zone.domains.zone_id
  depends_on      = [aptible_aws_acm.cert]
}

resource "aptible_aws_ecs_web" "web" {
  environment_id      = var.environment_id
  organization_id     = var.organization_id
  vpc_name            = aptible_aws_vpc.network.name
  depends_on          = [time_sleep.wait_30_seconds]

  name                = "nginxapp"
  container_name      = "nginxapp"
  container_image     = "nginx"
  lb_cert_arn         = aptible_aws_acm.cert.arn
  lb_cert_domain      = aptible_aws_acm.cert.fqdn

  is_public           = true # optional
  container_port      = 5000 # optional
  environment_secrets = {}
}

resource "aws_route53_record" "www" {
  zone_id = data.aws_route53_zone.domains.zone_id
  name    = aptible_aws_acm.cert.fqdn
  type    = "CNAME"
  ttl     = 300
  records = [aptible_aws_ecs_web.web.load_balancer_url]
}
