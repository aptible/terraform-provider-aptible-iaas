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

variable "secrets" {
  type      = map(string)
  sensitive = true
  default = {
    pass    = "123"
  }
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

resource "aptible_aws_secret" "secrets" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = "nextsecrets"
  secret_string   = jsonencode(var.secrets)
}

resource "aptible_aws_vpc" "network" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = "nextvpc" # optional
}

resource "aptible_aws_rds" "db" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name        = aptible_aws_vpc.network.name

  name            = "nextdb" # force new
  engine          = "postgres" # force new
  engine_version  = "14" # force new
}

resource "aptible_aws_redis" "cache" {
  environment_id      = data.aptible_environment.env.id
  organization_id     = data.aptible_organization.org.id
  vpc_name            = aptible_aws_vpc.network.name

  name                = "nextcache"

  description         = "integration testing" # optional
  snapshot_window     = "00:00-01:00" # optional
  maintenance_window  = "sun:10:00-sun:14:00" # optional
}

resource "aptible_aws_acm" "cert" {
  environment_id    = data.aptible_environment.env.id
  organization_id   = data.aptible_organization.org.id

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

resource "aptible_aws_acm_waiter" "waiter" {
  certificate_arn   = aptible_aws_acm.cert.arn
  validation_fqdns  = [for dns in local.validation_dns: dns.record] # optional
}

resource "aws_route53_record" "domains" {
  allow_overwrite = true
  name            = local.validation_dns.0.name
  records         = [local.validation_dns.0.record]
  ttl             = 60
  type            = local.validation_dns.0.type
  zone_id         = data.aws_route53_zone.domains.zone_id
}

resource "aptible_aws_ecs_web" "web" {
  environment_id      = data.aptible_environment.env.id
  organization_id     = data.aptible_organization.org.id
  vpc_name            = aptible_aws_vpc.network.name
  depends_on          = [aptible_aws_acm_water.waiter]

  name                = "nextapp"
  container_name      = "nextapp"
  container_image     = "quay.io/aptible/deploy-demo-app"
  lb_cert_arn         = aptible_aws_acm.cert.arn
  lb_cert_domain      = aptible_aws_acm.cert.fqdn

  connects_to       = [aptible_aws_rds.db.id, aptible_aws_redis.cache.id] # optional, for connecting to other resources
  is_public           = true # optional
  container_command   = [
    "gunicorn",
    "app:app",
    "-b",
    "0.0.0.0:5000",
    "--access-logfile",
    "-"] # optional
  container_port      = 5000 # optional
  environment_secrets = { # optional
    PASS = {
      secret_arn      = aptible_aws_secret.secrets.arn
      secret_json_key = "pass"
    }
    DATABASE_URL = {
      secret_arn      = aptible_aws_rds.db.uri_secret_arn
      secret_json_key = ""
    }
    REDIS_URL = {
      secret_arn      = aptible_aws_redis.cache.uri_secret_arn,
      secret_json_key = "dsn"
    }
    # TOKEN_SECRET = {
    #   secret_arn      = aptible_aws_redis.cache.uri_secret_arn,
    #   secret_kms_arn  = aptible_aws_redis.cache.secrets_kms_key_arn,
    #   secret_json_key = "token"
    # }
  }
}

resource "aptible_aws_ecs_compute" "worker" {
  environment_id      = data.aptible_environment.env.id
  organization_id     = data.aptible_organization.org.id
  vpc_name            = aptible_aws_vpc.network.name

  name                = "nextworker"
  container_name      = "nextworker"
  container_image     = "quay.io/aptible/deploy-demo-app"

  connects_to       = [aptible_aws_rds.db.id, aptible_aws_redis.cache.id] # optional, for connecting to other resources
  container_command   = ["python", "-m", "worker"] # optional
  container_port      = 5001 # optional
  environment_secrets = { # optional
    PASS = {
      secret_arn      = aptible_aws_secret.secrets.arn
      secret_json_key = "pass"
    }
    DATABASE_URL = {
      secret_arn      = aptible_aws_rds.db.uri_secret_arn
      secret_json_key = ""
    }
    REDIS_URL = {
      secret_arn      = aptible_aws_redis.cache.uri_secret_arn,
      secret_json_key = "dsn"
    }
  }
}

resource "aws_route53_record" "www" {
  zone_id = data.aws_route53_zone.domains.zone_id
  name    = aptible_aws_acm.cert.fqdn
  type    = "CNAME"
  ttl     = 300
  records = [aptible_aws_ecs_web.web.load_balancer_url]
}
