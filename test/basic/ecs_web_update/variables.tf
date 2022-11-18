variable "dns_account_id" {
  type = string
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

variable "vpc_name" {
  type = string
}

variable "domain" {
  type = string
}

variable "subdomain" {
  type = string
}

variable "container_image" {
  type = string
}

variable "container_command" {
  type = list(string)
}

variable "container_port" {
  type = number
}

variable "ecs_name" {
  type = string
}

variable "container_name" {
  type = string
}

variable "is_public" {
  type = bool
}

variable "is_ecr_image" {
  type = bool
}

variable "registry_credentials_arn" {
  type    = string
  default = null
}

variable "environment_secrets" {
  type    = map(map(string))
  default = {}
}