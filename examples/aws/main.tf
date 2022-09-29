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
  type = string
  default = "2253ae98-d65a-4180-aceb-8419b7416677"
}

variable "environment_id" {
  type = string
  default = "238930f4-0750-4f55-b43c-e1a11c437e23"
}

data "aptible_organization" "org" {
  id = var.organization_id
}

data "aptible_environment" "env" {
  id = var.environment_id
  org_id = data.aptible_organization.org.id
}

resource "aptible_aws_vpc" "network" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  asset_version   = "latest"
  name            = "my_vpc"
}

resource "aptible_aws_rds" "db" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name        = aptible_aws_vpc.network.name

  asset_version   = "latest"
  name            = "dbnext" # force new
  engine          = "postgres"
  engine_version  = "14"
}

resource "aptible_aws_redis" "cache" {
  environment_id = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name = aptible_aws_vpc.network.name

  asset_version  = "latest"
  name = "appcache"
  description = "integration testing"
  snapshot_window = "00:00-01:00"
  maintenance_window = "sun:10:00-sun:14:00"
}

resource "aptible_aws_acm" "cert" {
  environment_id = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id

  asset_version  = "latest"
  fqdn = "eric.aptible-test-leeroy.com"
  validation_method = "DNS"
}

data "aws_route53_zone" "example" {
  name         = "aptible-test-leeroy.com"
  private_zone = false
}

locals {
  validation_dns = {
    for dvo in aptible_aws_acm.cert.domain_validation_records : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }
}


resource "aws_route53_record" "example" {
  allow_overwrite = true
  name            = local.validation_dns.0.name
  records         = [local.validation_dns.0.record]
  ttl             = 60
  type            = local.validation_dns.0.type
  zone_id         = data.aws_route53_zone.example.zone_id
}

resource "time_sleep" "wait_30_seconds" {
  depends_on = [aws_route53_record.example]

  create_duration = "30s"
}

#resource "aws_route53_record" "example" {
#  for_each = {
#    for dvo in aptible_aws_acm.cert.domain_validation_records : dvo.domain_name => {
#      name   = dvo.resource_record_name
#      record = dvo.resource_record_value
#      type   = dvo.resource_record_type
#    }
#  }
#
#  allow_overwrite = true
#  name            = each.value.name
#  records         = [each.value.record]
#  ttl             = 60
#  type            = each.value.type
#  zone_id         = data.aws_route53_zone.example.zone_id
#}
#
#resource "time_sleep" "wait_30_seconds" {
#  depends_on = [aws_route53_record.example]
#
#  create_duration = "30s"
#}

resource "aptible_aws_ecs_web" "web" {
  environment_id = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name = aptible_aws_vpc.network.name

  asset_version  = "latest"
  name = "myapp"
  is_public = true
  container_name = "myapp"
  container_image = "quay.io/aptible/deploy-demo-app"
  container_command = ["echo", "'hi there'"]
  container_port = 5000
  lb_cert_arn = aptible_aws_acm.cert.arn
  lb_cert_domain = aptible_aws_acm.cert.fqdn
  environment_secrets = {
    DATABASE_URL = {
      secret_arn = aptible_aws_rds.db.uri_secret_arn
      secret_kms_arn = aptible_aws_rds.db.secrets_kms_key_arn,
      secret_json_key = ""
    }
    # TOKEN_SECRET = {
    #   secret_arn = aptible_aws_redis.cache.uri_secret_arn,
    #   secret_kms_arn = aptible_aws_redis.cache.secrets_kms_key_arn,
    #   secret_json_key = "token"
    # }
    REDIS_URL = {
      secret_arn = aptible_aws_redis.cache.uri_secret_arn,
      secret_kms_arn = aptible_aws_redis.cache.secrets_kms_key_arn,
      secret_json_key = "dsn"
    }
  }
  depends_on = [time_sleep.wait_30_seconds]
}

#resource "aptible_connection" "web_to_rds" {
#  inbound_id = aptible_aws_ecs_web.web.id
#  outbound_id = aptible_aws_rds.cache.id
#  inbound_label = "web"
#  outbound_label = "rds"
#}
#
#resource "aptible_connection" "web_to_redis" {
#  inbound_id = aptible_aws_ecs_web.web.id
#  outbound_id = aptible_aws_redis.db.id
#  inbound_label = "web"
#  outbount_label = "redis"
#}
