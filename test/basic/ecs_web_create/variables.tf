variable "aws_dns_role" {
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
