terraform {
  required_providers {
    aptible = {
      source  = "aptible.com/aptible/aptible-iaas"
    }
  }
}

variable "org_id" {
  type    = string
}

variable "env_id" {
  type    = string
}

variable "aptible_host" {
  type    = string
}

variable "fqdn" {
  type    = string
}

provider "aptible" {
  host = var.aptible_host
}

resource "aptible_aws_vpc" "network" {
  environment_id  = var.env_id
  organization_id = var.org_id
  asset_version   = "v0.26.1"
  name            = "conn" # optional
}

resource "aptible_aws_acm" "cert" {
  environment_id    = var.env_id
  organization_id   = var.org_id

  asset_version     = "v0.26.1"
  fqdn              = var.fqdn

  validation_method = "DNS" # optional
}

resource "aptible_aws_ecs_web" "web" {
  environment_id      = var.org_id
  organization_id     = var.env_id
  vpc_name            = aptible_aws_vpc.network.name

  asset_version       = "v0.26.1"
  name                = "nginx"
  container_name      = "nginx"
  container_image     = "nginx/alpine"
  lb_cert_arn         = aptible_aws_acm.cert.arn
  lb_cert_domain      = aptible_aws_acm.cert.fqdn

  connects_to         = ["321", "123", "222"]
  is_public           = true # optional
  container_command   = ["nginx", "-g", "daemon off;"] # optional
  container_port      = 80 # optional
  environment_secrets = {} # optional
}
