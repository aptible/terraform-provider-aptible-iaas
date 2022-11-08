terraform {
  required_providers {
    aptible = {
      source = "aptible.com/aptible/aptible-iaas"
    }
  }
}

variable "organization_id" {
  type = string
}

variable "environment_id" {
  type = string
}

variable "aptible_host" {
  type = string
}

variable "subdomain" {
  type = string
}

variable "domain" {
  type = string
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


resource "aptible_aws_acm" "cert" {
  environment_id  = data.aptible_environment.env.id
  organization_id = data.aptible_organization.org.id

  fqdn = "${var.subdomain}.${var.domain}"

  validation_method = "DNS"
}

output "cert_id" {
  value = aptible_aws_acm.cert.id
}

output "domain_validation_records" {
  value = aptible_aws_acm.cert.domain_validation_records
}

output "cert_arn" {
  value = aptible_aws_acm.cert.arn
}

output "fqdn" {
  value = aptible_aws_acm.cert.fqdn
}


