terraform {
  required_providers {
    aptible = {
      source  = "aptible.com/aptible/aptible-iaas"
    }
  }
}

provider "aptible" {
  host = "cloud-api.cloud.aptible.com"
}

variable "organization_id" {
  type = string
}

variable "environment_id" {
  type = string
}

data "aptible_organization" "org" {
  # id = "2253ae98-d65a-4180-aceb-8419b7416677"
  id = var.organization_id
}

data "aptible_environment" "env" {
  # id = "238930f4-0750-4f55-b43c-e1a11c437e23"
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
  environment_id = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name = aptible_aws_vpc.network.name

  asset_version  = "latest"
  name = "app_redis"
  engine = "postgres"
  version = "13"
}

#resource "aptible_aws_redis" "cache" {
#  environment_id = data.aptible_environment.env.id
#  organization_id = data.aptible_organization.org.id
#  vpc_name = aptible_aws_vpc.network.name
#
#  asset_version  = "latest"
#  name = "app_redis"
#}
#
#resource "aptible_aws_acm" "cert" {
#  environment_id = data.aptible_environment.env.id
#  organization_id = data.aptible_organization.org.id
#
#  asset_version  = "latest"
#  fqdn = "www.example.com"
#  validation_method = "EMAIL"
#}
#
#resource "aptible_aws_ecs_web" "web" {
#  environment_id = data.aptible_environment.env.id
#  organization_id = data.aptible_organization.org.id
#  vpc_name = aptible_asset.network.name
#
#  asset_version  = "latest"
#  name = "my_app"
#  is_public = true
#  container_name = "my_app"
#  container_image = "quay.io/aptible/deploy-demo-app"
#  container_port = 5000
#  lb_cert_arn = jsondecode(aptible_asset.cert.outputs).arn
#  environment_secrets = {
#    DATABASE_URL = {
#      secret_arn = "" // jsondecode(aptible_asset.rds.outputs).uri_arn
#    }
#    REDIS_URL = {
#      secret_arn = ""
#    }
#  }
#}
#
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
