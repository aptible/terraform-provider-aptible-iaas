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

data "aptible_aws_vpc" "network" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  name            = var.vpc_name
}

resource "aptible_aws_rds" "database" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id

  vpc_name       = data.aptible_aws_vpc.network.name
  name           = var.database_name
  engine         = var.database_engine
  engine_version = var.database_engine_version
}
