variable "organization_id" {
  type        = string
  description = "Your Aptible organization ID"
}

variable "environment_id" {
  type        = string
  description = "Your Aptible environment ID"
}

variable "domain" {
  type        = string
  description = "Parent domain (e.g. 'customer.com' for deploying onto demo-app.customer.com)"

}

variable "subdomain" {
  type        = string
  description = "Subdomain (e.g. 'example' for deploying onto demo-app.customer.com)"
}

variable "vpc_name" {
  type        = string
  description = "Name for the VPC containing ECS/RDS/ElastiCache"
  default     = "demo-vpc"
}

variable "container_image" {
  type        = string
  description = "Docker image to deploy"
  default     = "quay.io/aptible/deploy-demo-app:latest"
}

variable "container_web_command" {
  type        = list
  description = "Array of arguments for the container's web service command"
  default     = ["gunicorn", "app:app", "-b", "0.0.0.0:5000", "--access-logfile", "-"]
}

variable "container_port" {
  type        = number
  description = "Port on which container is listening for HTTP"
  default     = 5000
}
