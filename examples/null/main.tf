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

resource "aptible_null_simple" "network" {
  environment_id          = data.aptible_environment.env.id
  organization_id         = data.aptible_organization.org.id

  asset_type = "simple"
  asset_platform = "null"
  asset_version  = "latest"

  name                   = "my_null"
}
