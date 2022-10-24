terraform {
  required_providers {
    aptible = {
      source  = "aptible.com/aptible/aptible-iaas"
    }
  }
}

variable "aptible_host" {
  type    = string
}

variable "organization_id" {
  type    = string
}

variable "environment_id" {
  type    = string
}

variable "vpc_name" {
  type = string
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

data "aptible_aws_vpc" "main" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name      = var.vpc_name
}

resource "aptible_aws_rds" "db" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name        = data.aptible_aws_vpc.main.name

  name            = "example-1"
  engine          = "postgres"
  engine_version  = "14"
}