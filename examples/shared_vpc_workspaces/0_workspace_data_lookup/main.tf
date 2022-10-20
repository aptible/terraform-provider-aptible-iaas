terraform {
  required_providers {
    aptible = {
      source  = "aptible.com/aptible/aptible-iaas"
    }
  }
}

variable "aptible_host" {
  type    = string
}

variable "organization_id" {
  type    = string
}

variable "environment_id" {
  type    = string
}

variable "vpc_id" {
  type = string
}

provider "aptible" {
  host = var.aptible_host
}

data "aptible_organization" "org" {
  id = var.organization_id
}

data "aptible_environment" "env" {
  id      = var.environment_id
  org_id  = data.aptible_organization.org.id
}

data "aptible_aws_vpc" "main" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id
  id              = var.vpc_id
}
