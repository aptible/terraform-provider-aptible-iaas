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
  token = var.aptible_token
}

data "aptible_organization" "org" {
  id = var.organization_id
}

data "aptible_environment" "env" {
  id      = var.environment_id
  org_id  = data.aptible_organization.org.id
}

resource "aptible_aws_vpc" "network" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  asset_version   = "latest"
  name            = var.vpc_name
}

resource "aptible_aws_rds" "db" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name        = aptible_aws_vpc.network.name
  depends_on      = [aptible_aws_vpc.network]

  asset_version   = "latest"
  name            = "demo-postgres"
  engine          = "postgres"
  engine_version  = "14"
}

resource "aptible_aws_redis" "cache" {
  environment_id      = data.aptible_environment.env.id
  organization_id     = data.aptible_organization.org.id
  vpc_name            = aptible_aws_vpc.network.name
  depends_on          = [aptible_aws_vpc.network]

  asset_version       = "latest"
  name                = "demo-redis"
  description         = "integration testing"
  snapshot_window     = "00:00-01:00"
  maintenance_window  = "sun:10:00-sun:14:00"
}

resource "aptible_aws_acm" "cert" {
  environment_id    = data.aptible_environment.env.id
  organization_id   = data.aptible_organization.org.id

  asset_version     = "latest"
  fqdn              = join(".", [var.subdomain, var.domain])
  validation_method = "DNS"
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

# TODO: replace with ACM waiter
# See https://github.com/aptible/terraform-aws-core/pull/114
resource "time_sleep" "wait_30_seconds" {
  depends_on      = [aws_route53_record.domains]
  create_duration = "30s"
}

resource "aptible_aws_ecs_web" "web" {
  environment_id      = data.aptible_environment.env.id
  organization_id     = data.aptible_organization.org.id
  vpc_name            = aptible_aws_vpc.network.name
  depends_on          = [time_sleep.wait_30_seconds]

  asset_version       = "latest"
  name                = "demo-app"
  is_public           = true
  container_name      = "demo-app-web"
  container_image     = var.container_image
  container_command   = var.container_web_command
  container_port      = var.container_port
  lb_cert_arn         = aptible_aws_acm.cert.arn
  lb_cert_domain      = aptible_aws_acm.cert.fqdn
  environment_secrets = {
    DATABASE_URL = {
      secret_arn      = aptible_aws_rds.db.uri_secret_arn
      secret_kms_arn  = aptible_aws_rds.db.secrets_kms_key_arn,
      secret_json_key = ""
    }
    REDIS_URL = {
      secret_arn      = aptible_aws_redis.cache.uri_secret_arn,
      secret_kms_arn  = aptible_aws_redis.cache.secrets_kms_key_arn,
      secret_json_key = "dsn"
    }
  }
}
