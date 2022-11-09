output "vpc_id" {
  value = aptible_aws_vpc.network.id
}

output "ecs_compute_id" {
  value = aptible_aws_ecs_compute.compute.id
}
