
output "vpc_id" {
  value = values(aptible_aws_vpc.vpc).id
}