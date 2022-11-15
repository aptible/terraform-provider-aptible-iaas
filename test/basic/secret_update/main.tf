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


resource "aptible_aws_secret" "secret" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id

  name          = var.secret_name
  secret_string = var.secret_value
}
