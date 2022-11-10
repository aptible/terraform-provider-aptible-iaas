terraform {
  required_providers {
    aptible = {
      source = "aptible.com/aptible/aptible-iaas"
    }
  }
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

resource "aptible_aws_vpc" "network" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = var.vpc_name
}

resource "aptible_aws_redis" "database" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id

  name        = var.node_name
  description = var.description
  vpc_name    = aptible_aws_vpc.network.name

  maintenance_window = var.maintenance_window
  snapshot_window    = var.snapshot_window
}


