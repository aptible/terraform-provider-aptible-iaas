terraform {
  required_providers {
    aptible = {
      source  = "aptible.com/aptible/aptible-iaas"
    }
  }
}

provider "aptible" {
  host = "cloud-api.cloud.aptible.com"
}

data "aptible_organization" "org" {
  id = "2253ae98-d65a-4180-aceb-8419b7416677"
}

data "aptible_environment" "env" {
  id = "238930f4-0750-4f55-b43c-e1a11c437e23"
  org_id = data.aptible_organization.org.id
}

resource "aptible_aws_vpc" "network" { // MADHU
  environment_id          = data.aptible_environment.env.id // TODO - make this a resource
  organization_id         = data.aptible_organization.org.id // TODO - make this a resource

  asset_type = "vpc"
  asset_platform = "aws"
  asset_version  = "latest"

  name                   = "my_vpc"
  #  cidr_block             = "10.43.0.0/16"
  #  max_availability_zones = 3
  #  nat_type               = "instance"
}

#resource "aptible_aws_rds" "db" { // MADHU
#  environment_id          = data.aptible_environment.env.id // TODO - make this a resource
#  organization_id         = data.aptible_organization.org.id // TODO - make this a resource
#
#  asset_type = "rds"
#  asset_platform = "aws"
#  asset_version  = "latest"
#
#  vpc_name       = aptible_aws_vpc.network.name
#  name           = "prod_database"
#  engine         = "mysql"
#  engine_version = "5.7"
#  pit_identifier = "super_unique_point_in_time_identifier"
#  tags = {}
#}

#resource "aptible_aws_ecs" "app" { // MADHU
#  environment_id          = data.aptible_environment.env.id
#  organization_id         = data.aptible_organization.org.id
#
#  asset_type = "ecs"
#  asset_platform = "aws"
#  asset_version  = "latest"
#
#  vpc_name       = aptible_aws_vpc.network.name
#  name           = "prod_ecs"
#  tags = {}
#}
#
#resource "aptible_resource_connection" "database_to_ecs" {
#  name = "ecs_to_rds"
#  description = "Allow ECS to Access RDS Database"
#  incoming_asset_connection = aptible_asset.db.id
#  outgoing_asset_connection = aptible_asset.app.id
#}
