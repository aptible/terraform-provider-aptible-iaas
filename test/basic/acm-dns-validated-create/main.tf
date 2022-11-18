terraform {
  required_providers {
    aptible = {
      source = "aptible.com/aptible/aptible-iaas"
    }
    aws = {
      source = "hashicorp/aws"
    }
  }
}

variable "aws_dns_role" {
  type = string
}
provider "aws" {
  alias  = "dns_account"
  region = "us-east-1"
  assume_role {
    role_arn = var.aws_dns_role
  }
}

variable "organization_id" {
  type = string
}

variable "environment_id" {
  type = string
}

variable "aptible_host" {
  type = string
}

variable "subdomain" {
  type = string
}

variable "domain" {
  type = string
}

provider "aptible" {
  host = var.aptible_host
}

data "aptible_organization" "org" {
  id = var.organization_id
}

data "aptible_environment" "env" {
  id     = var.environment_id
  org_id = data.aptible_organization.org.id
}

resource "aptible_aws_acm" "cert" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id

  fqdn = "${var.subdomain}.${var.domain}"

  validation_method = "DNS"
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

data "aws_route53_zone" "domains" {
  name         = var.domain
  private_zone = false
  provider     = aws.dns_account
}

resource "aws_route53_record" "domains" {
  allow_overwrite = true
  name            = local.validation_dns.0.name
  records         = [local.validation_dns.0.record]
  ttl             = 60
  type            = local.validation_dns.0.type
  zone_id         = data.aws_route53_zone.domains.zone_id
  depends_on      = [aptible_aws_acm.cert]
  provider        = aws.dns_account
}

resource "aptible_aws_acm_waiter" "waiter" {
  environment_id   = data.aptible_environment.env.id
  organization_id  = data.aptible_organization.org.id
  certificate_arn  = aptible_aws_acm.cert.arn
  validation_fqdns = [for dns in local.validation_dns : dns.name]
}

output "cert_id" {
  value = aptible_aws_acm.cert.id
}

output "cert_waiter_id" {
  value = aptible_aws_acm_waiter.waiter.id
}

output "domain_validation_records" {
  value = aptible_aws_acm.cert.domain_validation_records
}

output "cert_arn" {
  value = aptible_aws_acm.cert.arn
}

output "fqdn" {
  value = aptible_aws_acm.cert.fqdn
}

output "aptible_aws_account_id" {
  value = data.aptible_environment.env.aws_account_id
}
