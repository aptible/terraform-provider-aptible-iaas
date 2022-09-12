terraform {
  required_providers {
    aptible = {
      source  = "aptible/aptible-iaas"
      version = "~>0.1"
    }
  }
}

data "aptible_organization" "org" { // ERIC
  id = "<UUID>" // TODO - revisit
}

data "aptible_environment" "env" { // ERIC
  id = "<UUID>" // TODO - revisit
  org_id = data.org.id
}

resource "aptible_aws_vpc" "network" { // MADHU
  environment_id         = data.env.id // TODO - make this a resource

  asset_version          = "1.3"
  name                   = "my_vpc"
  cidr_block             = "10.43.0.0/16"
  max_availability_zones = 3
  nat_type               = "instance"
}

resource "aptible_aws_rds" "db" { // MADHU
  environment_id = data.env.id // TODO - make this a resource

  asset_version  = "1.3"
  vpc_name       = aptible_aws_vpc.network.name
  name           = "prod_database"
  engine         = "mysql"
  engine_version = "5.7"
  pit_identifier = "super_unique_point_in_time_identifier"
  tags = {}
}

resource "aptible_aws_ecs" "app" { // MADHU
  environment_id = data.env.id // TODO - make this a resource

  asset_version  = "1.3"
  vpc_name       = aptible_aws_vpc.network.name
  name           = "prod_ecs"
  tags = {}
}

resource "aptible_resource_connection" "database_to_ecs" { // TODO
  name = "ecs_to_rds"
  description = "Allow ECS to Access RDS Database"
  incoming_asset_connection = aptible_aws_rds.db.id
  outgoing_asset_connection = aptible_aws_ecs.app.id
  tags = {}
}
