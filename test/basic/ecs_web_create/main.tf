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

provider "aws" {
  alias  = "dns_account"
  region = "us-east-1"
  assume_role {
    role_arn = "arn:aws:iam::${var.dns_account_id}:role/OrganizationAccountAccessRole"
  }
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

resource "aptible_aws_vpc" "vpc" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = var.vpc_name
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

resource "aptible_aws_ecs_web" "web" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name        = aptible_aws_vpc.vpc.name
  depends_on      = [aptible_aws_acm_waiter.waiter]

  name                  = var.ecs_name
  container_name        = var.container_name
  container_image       = var.container_image
  wait_for_steady_state = true

  container_command   = var.container_command
  container_port      = var.container_port
  is_public           = var.is_public
  is_ecr_image        = var.is_ecr_image
  lb_cert_arn         = aptible_aws_acm.cert.arn
  lb_cert_domain      = aptible_aws_acm.cert.fqdn
  environment_secrets = {}
}
