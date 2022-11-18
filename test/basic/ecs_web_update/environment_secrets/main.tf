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

variable "plain_secret_1_name" {
  type = string
}

variable "plain_secret_1" {
  type = string
}

variable "plain_secret_2_name" {
  type = string
}

variable "plain_secret_2" {
  type = string
}

variable "json_secrets" {
  type = map(string)
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

resource "aptible_aws_secret" "plain_secret_1" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = var.plain_secret_1_name
  secret_string   = var.plain_secret_1
}

resource "aptible_aws_secret" "plain_secret_2" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = var.plain_secret_2
  secret_string   = var.plain_secret_2
}

resource "aptible_aws_secret" "json_secrets" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = "jsonsecrets"
  secret_string   = jsonencode(var.json_secrets)
}

output "plain_secret_1_arn" {
  value = aptible_aws_secret.plain_secret_1.arn
}

output "plain_secret_2_arn" {
  value = aptible_aws_secret.plain_secret_1.arn
}

output "json_secret_arn" {
  value = aptible_aws_secret.json_secrets.arn
}