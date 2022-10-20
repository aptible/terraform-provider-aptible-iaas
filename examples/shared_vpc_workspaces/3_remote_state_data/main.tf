data "terraform_remote_state" "vpc" {
  # there are other backends possible, but they must point to the
  # workspace with the vpc asset
  backend = "local"

  config = {
    path = "../1_workspace_with_vpc_resource"
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
  env_id    = data.aptible_environment.id
  org_id    = data.aptible_organization.org.id
  name      = data.terraform_remote_state.vpc.name
  # id
  # name
}

resource "aptible_aws_rds" "db" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name        = data.aptible_aws_vpc.main.name

  name            = "nextdb" # force new
  engine          = "postgres" # force new
  engine_version  = "14" # force new
}