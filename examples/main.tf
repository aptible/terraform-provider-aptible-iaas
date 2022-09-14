terraform {
  required_providers {
    aptible = {
      source  = "aptible.com/aptible/aptible-iaas"
    }
  }
}

data "aptible_organization" "org" { // ERIC
  id = "<UUID>" // TODO - revisit
}

data "aptible_environment" "env" { // ERIC
  id = "<UUID>" // TODO - revisit
  org_id = data.aptible_organization.org.id
}

resource "aptible_asset" "network" { // MADHU
  environment_id          = data.aptible_environment.env.id // TODO - make this a resource
  organization_id         = data.aptible_organization.org.id // TODO - make this a resource

  asset_type = "vpc"
  asset_platform = "aws"
  asset_version  = "latest"

  parameters = jsonencode({
    name                   = "my_vpc"
    cidr_block             = "10.43.0.0/16"
    max_availability_zones = 3
    nat_type               = "instance"
  })
}

resource "aptible_asset" "db" { // MADHU
  environment_id          = data.aptible_environment.env.id // TODO - make this a resource
  organization_id         = data.aptible_organization.org.id // TODO - make this a resource

  asset_type = "rds"
  asset_platform = "aws"
  asset_version  = "latest"
  parameters = jsonencode({
    vpc_name       = jsondecode(aptible_asset.network.parameters).name
    name           = "prod_database"
    engine         = "mysql"
    engine_version = "5.7"
    pit_identifier = "super_unique_point_in_time_identifier"
    tags = {}
  })
}

resource "aptible_asset" "app" { // MADHU
  environment_id          = data.aptible_environment.env.id // TODO - make this a resource
  organization_id         = data.aptible_organization.org.id // TODO - make this a resource

  asset_type = "ecs"
  asset_platform = "aws"
  asset_version  = "latest"
  parameters = jsonencode({
    vpc_name       = jsondecode(aptible_asset.network.parameters).name
    name           = "prod_ecs"
    tags = {}
  })
}
#
resource "aptible_resource_connection" "database_to_ecs" { // TODO
  name = "ecs_to_rds"
  description = "Allow ECS to Access RDS Database"
  incoming_asset_connection = aptible_asset.db.id
  outgoing_asset_connection = aptible_asset.app.id
}
