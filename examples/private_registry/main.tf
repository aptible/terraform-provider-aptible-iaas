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

provider "aws" {
  region = "us-east-1"
}

provider "aptible" {
  host = "cloud-api.cloud.aptible.com"
}

variable "organization_id" {
  type    = string
}

variable "environment_id" {
  type    = string
}

variable "secret_registry" {
  type    = map(string)
  default = {
    username = ""
    password = ""
  }
}

data "aptible_organization" "org" {
  id = var.organization_id
}

data "aptible_environment" "env" {
  id      = var.environment_id
  org_id  = data.aptible_organization.org.id
}

resource "aptible_aws_secret" "registry" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  asset_version   = "v0.26.1"
  name            = "registry"
  secret_string   = jsonencode(var.secret_registry)
}

resource "aptible_aws_vpc" "network" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  asset_version   = "v0.26.1"
  name            = "priv" # optional
}

resource "aptible_aws_acm" "cert" {
  environment_id    = data.aptible_environment.env.id
  organization_id   = data.aptible_organization.org.id

  asset_version     = "v0.26.1"
  fqdn              = "erock.aptible-test-leeroy.com"

  validation_method = "DNS" # optional
}

data "aws_route53_zone" "domains" {
  name         = "aptible-test-leeroy.com"
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
  name                = "privimg"
  container_name      = "privimg"
  container_image     = "ghcr.io/aptible/docker-hello-world-private:main"
  lb_cert_arn         = aptible_aws_acm.cert.arn
  lb_cert_domain      = aptible_aws_acm.cert.fqdn

  container_registry_secret_arn = aptible_aws_secret.registry.arn # optional
  is_public           = true # optional
  container_command   = ["nginx", "-g", "daemon off;"] # optional
  container_port      = 80 # optional
  environment_secrets = {} # optional
}

resource "aws_route53_record" "www" {
  zone_id = data.aws_route53_zone.domains.zone_id
  name    = aptible_aws_acm.cert.fqdn
  type    = "CNAME"
  ttl     = 300
  records = [aptible_aws_ecs_web.web.load_balancer_url]
}
