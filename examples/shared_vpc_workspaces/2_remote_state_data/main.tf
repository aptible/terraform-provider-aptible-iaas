terraform {
  required_providers {
    aptible = {
      source  = "aptible.com/aptible/aptible-iaas"
    }
  }
}

data "terraform_remote_state" "vpc" {
  # there are other backends possible, but they must point to the
  # workspace with the vpc asset
  backend = "local"

  config = {
    path = "../0_workspace_with_vpc_resource/terraform.tfstate"
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
  environment_id    = data.aptible_environment.env.id
  organization_id    = data.aptible_organization.org.id
  name        = data.terraform_remote_state.vpc.outputs.vpc_name
}

resource "aptible_aws_rds" "db" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name        = data.aptible_aws_vpc.main.name

  name            = "example-2"
  engine          = "postgres"
  engine_version  = "14"
}