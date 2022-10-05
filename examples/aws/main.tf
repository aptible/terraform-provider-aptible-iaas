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
  default = "2253ae98-d65a-4180-aceb-8419b7416677"
}

variable "environment_id" {
  type    = string
  default = "b47357cd-2971-4f73-ad6f-2edbddcde529"
}

# -OR-
# variable "secret_pass" {
#   type    = string
#   default = "123"
# }

variable "secrets" {
  type      = map(string)
  sensitive = true
  default = {
    pass    = "123"
    other   = "abc"
  }
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
  asset_version   = "latest"
  name            = "mysecrets"
  secret_string   = jsonencode(var.secrets)
  # secret_string   = var.secret_pass
}

resource "aptible_aws_vpc" "network" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  asset_version   = "latest"
  name            = "myvpc"
}

resource "aptible_aws_rds" "db" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name        = aptible_aws_vpc.network.name
  depends_on      = [aptible_aws_vpc.network]

  asset_version   = "latest"
  name            = "mydb" # force new
  engine          = "postgres"
  engine_version  = "14"
}

resource "aptible_aws_redis" "cache" {
  environment_id      = data.aptible_environment.env.id
  organization_id     = data.aptible_organization.org.id
  vpc_name            = aptible_aws_vpc.network.name
  depends_on          = [aptible_aws_vpc.network]

  asset_version       = "latest"
  name                = "mycache"
  description         = "integration testing"
  snapshot_window     = "00:00-01:00"
  maintenance_window  = "sun:10:00-sun:14:00"
}

resource "aptible_aws_acm" "cert" {
  environment_id    = data.aptible_environment.env.id
  organization_id   = data.aptible_organization.org.id

  asset_version     = "latest"
  fqdn              = "eric.aptible-test-leeroy.com"
  validation_method = "DNS"
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

  asset_version       = "latest"
  name                = "myapp"
  is_public           = true
  container_name      = "myapp"
  container_image     = "quay.io/aptible/deploy-demo-app"
  container_command   = ["python", "-m", "gunicorn", "app:app", "-b", "0.0.0.0:5000", "--access-logfile", "-"]
  container_port      = 5000
  lb_cert_arn         = aptible_aws_acm.cert.arn
  lb_cert_domain      = aptible_aws_acm.cert.fqdn
  environment_secrets = {
    PASS = {
      secret_arn      = aptible_aws_secret.secrets.arn
      secret_kms_arn  = aptible_aws_secret.secrets.kms_arn,
      secret_json_key = "pass"
      # secret_json_key = ""
    }
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

  asset_version       = "latest"
  name                = "myworker"
  container_name      = "myworker"
  container_image     = "quay.io/aptible/deploy-demo-app"
  container_command   = ["python", "-m", "worker"]
  container_port      = 5001
  environment_secrets = {
    PASS = {
      secret_arn      = aptible_aws_secret.secrets.arn
      secret_kms_arn  = aptible_aws_secret.secrets.kms_arn,
      secret_json_key = "pass"
      # secret_json_key = ""
    }
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
    # TOKEN_SECRET = {
    #   secret_arn      = aptible_aws_redis.cache.uri_secret_arn,
    #   secret_kms_arn  = aptible_aws_redis.cache.secrets_kms_key_arn,
    #   secret_json_key = "token"
    # }
  }
}
