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
}

data "aptible_organization" "org" {
  id = var.organization_id
}

data "aptible_environment" "env" {
  id      = var.environment_id
  org_id  = data.aptible_organization.org.id
}

resource "aptible_aws_secret" "service_secrets" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  asset_version   = "v0.26.1"
  name            = "nginxsecrets"
  secret_string   = jsonencode({
    "VARIABLE1": "vairable 1 secret",
    "VARIABLE2": "vairable 2 secret"
  })
}

resource "aptible_aws_secret" "special_service_secret" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  asset_version   = "v0.26.1"
  name            = "nginxsecret_special"
  secret_string   = "aspecialsecret"
}

resource "aptible_aws_vpc" "network" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  asset_version   = "v0.26.1"
  name            = "network123"
}

resource "aptible_aws_acm" "cert" {
  environment_id    = data.aptible_environment.env.id
  organization_id   = data.aptible_organization.org.id

  asset_version     = "v0.26.1"
  fqdn              = var.fqdn

  validation_method = "DNS"
}

data "aws_route53_zone" "domains" {
  name         = var.domain
  private_zone = false
}

locals {
  validation_dns = [
    for record in aptible_aws_acm.cert.domain_validation_records : {
      name   = record.resource_record_name
      record = record.resource_record_value
      type   = record.resource_record_type
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

resource "time_sleep" "wait_30_seconds" {
  depends_on      = [aws_route53_record.domains]
  create_duration = "30s"
}

resource "aptible_aws_ecs_web" "web" {
  environment_id      = data.aptible_environment.env.id
  organization_id     = data.aptible_organization.org.id
  vpc_name            = aptible_aws_vpc.network.name
  depends_on          = [time_sleep.wait_30_seconds]

  asset_version       = "v0.26.1"
  name                = "nginx"
  container_name      = "nginx"
  container_image     = "nginx:1.23"
  lb_cert_arn         = aptible_aws_acm.cert.arn
  lb_cert_domain      = aptible_aws_acm.cert.fqdn

  is_public           = true
  container_command   = ["nginx", "-g", "daemon off;"]
  container_port = 80
  environment_secrets = {
    "CONTAINER_VAR1": {
      "secret_arn": aptible_aws_secret.service_secrets.arn,
      "secret_json_key": "VARIABLE1"
    },
    "CONTAINER_VAR2": {
      "secret_arn": aptible_aws_secret.special_service_secret.arn
    },
  }
}

resource "aws_route53_record" "www" {
  zone_id = data.aws_route53_zone.domains.zone_id
  name    = aptible_aws_acm.cert.fqdn
  type    = "CNAME"
  ttl     = 300
  records = [aptible_aws_ecs_web.web.load_balancer_url]
}
