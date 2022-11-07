output "vpc_id" {
  value = aptible_aws_vpc.network.id
}

output "rds_id" {
  value = aptible_aws_rds.database.id
}