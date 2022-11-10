output "vpc_id" {
  value = aptible_aws_vpc.network.id
}

output "redis_id" {
  value = aptible_aws_redis.database.id
}

output "redis_uri_secret_arn" {
  value = aptible_aws_redis.database.uri_secret_arn
}

output "redis_secrets_kms_key_arn" {
  value = aptible_aws_redis.database.secrets_kms_key_arn
}

output "redis_arn" {
  value = aptible_aws_redis.database.elasticache_arn
}

output "cluster_id" {
  value = aptible_aws_redis.database.elasticache_cluster_id
}
