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