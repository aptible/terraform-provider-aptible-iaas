terraform {
  required_providers {
    aptible = {
      source  = "aptible.com/aptible/aptible-iaas"
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

variable "add_secret" {
  type = bool
}

variable "add_vpc" {
  type = bool
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
  for_each = var.add_secret ? toset(["ok"]) : []
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = "test_secrets"
  secret_string   = jsonencode(var.secrets)
}

resource "aptible_aws_vpc" "vpc" {
  for_each = var.add_vpc ? toset(["ok"]) : []
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = "test_vpc"
}

output "secrets_id" {
  value = values(aptible_aws_secret.secrets).*.id
}

output "vpc_id" {
  value = values(aptible_aws_vpc.vpc).*.id
}