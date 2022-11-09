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

variable "domain" {
  type        = string
  description = "Parent domain (e.g. 'customer.com' for deploying onto demo-app.customer.com)"
}

variable "subdomain" {
  type        = string
  description = "Subdomain (e.g. 'example' for deploying onto demo-app.customer.com)"
}

variable "container_image" {
  type        = string
  description = "Docker image to deploy"
  default     = "quay.io/aptible/deploy-demo-app:latest"
}

variable "container_web_command" {
  type        = list(string)
  description = "Array of arguments for the container's web service command"
  default     = ["gunicorn", "app:app", "-b", "0.0.0.0:5000", "--access-logfile", "-"]
}

variable "container_port" {
  type        = number
  description = "Port on which container is listening for HTTP"
  default     = 5000
}

variable "env_vars" {
  type      = map(string)
  sensitive = true
  default = {
    PUBLIC_VALUE = "123"
  }
}

variable "secrets" {
  type      = map(string)
  sensitive = true
  default = {
    PASSWORD = "123"
  }
}
