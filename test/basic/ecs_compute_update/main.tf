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

resource "aptible_aws_secret" "secret" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id

  name          = var.secret_name
  secret_string = var.secret_json
}

resource "aptible_aws_ecs_compute" "compute" {
  // setting this below depends on to force secret to be made first, there are some cases in testing
  // where we will need that secret value before compute
  depends_on = [aptible_aws_secret.secret]

  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  vpc_name        = aptible_aws_vpc.network.name

  name            = var.compute_name
  container_name  = "${var.compute_name}-container"
  container_image = var.container_image

  container_command   = var.container_command
  container_port      = 80
  environment_secrets = var.environment_secrets

  container_registry_secret_arn = var.container_registry_secret_arn
  is_ecr_image                  = var.is_ecr_image

  wait_for_steady_state = true
}
