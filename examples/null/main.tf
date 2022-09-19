terraform {
  required_providers {
    aptible = {
      source  = "aptible.com/aptible/aptible-iaas"
    }
  }
}

data "aptible_organization" "org" {
  id = "ORG_ID"
}

data "aptible_environment" "env" {
  id = "ENV_ID"
  org_id = data.aptible_organization.org.id
}

resource "aptible_null_simple" "network" { // MADHU
  environment_id          = data.aptible_environment.env.id
  organization_id         = data.aptible_organization.org.id

  asset_type = "simple"
  asset_platform = "null"
  asset_version  = "latest"

  name                   = "my_null"
}
