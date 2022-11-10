module "db" {
  source = "../"


  aptible_host            = var.aptible_host
  database_engine         = "aurora"
  database_engine_version = var.database_engine_version
  database_name           = var.database_name
  environment_id          = var.environment_id
  organization_id         = var.organization_id
  vpc_name                = var.vpc_name
}