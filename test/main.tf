terraform {
  required_providers {
    null = {
      version = "~> 3.0"
    }
    aptible = {
      source  = "aptible.com/aptible/aptible-iaas"
    }
  }
}

provider "aptible" {
  host = "cloud-api.sandbox.aptible-cloud-staging.com"
}

data "aptible_organization" "org" {
  id = "e6c7394d-054c-454f-9710-dc02fa7406d3"
}

resource "null_resource" "web" {
  provisioner "local-exec" {
    command = "echo ${data.aptible_organization.org.name}"
  }
}
