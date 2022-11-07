terraform {
  required_providers {
    aptible = {
      source = "aptible.com/aptible/aptible-iaas"
    }
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
  id     = var.environment_id
  org_id = data.aptible_organization.org.id
}

resource "aptible_aws_secret" "secrets" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = "test_secrets"
  secret_string   = jsonencode(var.secrets)
}

resource "aptible_aws_vpc" "vpc" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = "test_vpc"
}

output "secrets_id" {
  value = values(aptible_aws_secret.secrets).id
}

output "vpc_id" {
  value = values(aptible_aws_vpc.vpc).id
}