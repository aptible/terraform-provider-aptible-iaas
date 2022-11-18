output "vpc_id" {
  value = aptible_aws_vpc.network.id
}

output "rds_id" {
  value = aptible_aws_rds.database.id
}

output "rds_db_identifier" {
  value = aptible_aws_rds.database.db_identifier
}
