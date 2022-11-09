variable "organization_id" {
  type = string
}

variable "environment_id" {
  type = string
}

variable "aptible_host" {
  type = string
}

variable "compute_name" {
  type = string
}

variable "container_command" {
  type = list(string)
}

variable "container_port" {
  type = number
}

variable "container_image" {
  type = string
}

variable "vpc_name" {
  type = string
}

variable "is_ecr_image" {
  type = bool
}

variable "environment_secrets" {
  default = {}
  type    = map(any)
}

variable "container_registry_secret_arn" {
  default = ""
  type    = string
}

variable "secret_registry" {
  default = {}
  type    = map(any)
}

variable "secret_name" {
  type = string
}

variable "secret_json" {
  type = string
}
