terraform {
  required_providers {
    aptible = {
      source = "aptible.com/aptible/aptible-iaas"
    }
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

variable "registry_secret_name" {
  type = string
}

variable "registry_secret_value" {
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

resource "aptible_aws_secret" "registry_secret" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = var.registry_secret_name
  secret_string   = var.registry_secret_value
}

output "registry_secret_id" {
  value = aptible_aws_secret.registry_secret.id
}

output "registry_secret_arn" {
  value = aptible_aws_secret.registry_secret.arn
}