terraform {
  required_providers {
    null = {
      version = "~> 3.0"
    }
    aptible = {
      source  = "aptible.com/prod/aptible-iaas"
    }
  }
}

data "aptible_organization" "org" { // ERIC
  id = "5b68c91d-7301-421c-8e14-5ba629e2d5f0"
}

resource "null_resource" "web" {
  provisioner "local-exec" {
    command = "echo ${data.aptible_organization.org.name}"
  }
}
