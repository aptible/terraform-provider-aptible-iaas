package elasticache

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	legacy_aws_sdk_ec2 "github.com/aws/aws-sdk-go/service/ec2"
	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func cleanupAndAssert(t *testing.T, terraformOptions *terraform.Options) {
	terraform.Destroy(t, terraformOptions)

	// test / assert all failures here
}

func TestElasticacheUpdate(t *testing.T) {
	vpc_name := "test-elasticache-vpc"
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"organization_id":    os.Getenv("ORGANIZATION_ID"),
			"environment_id":     os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":       os.Getenv("APTIBLE_HOST"),
			"node_name":          "test-elasticache-node",
			"vpc_name":           vpc_name,
			"description":        "TestingModule",
			"snapshot_window":    "01:00-04:00",
			"maintenance_window": "sun:05:00-sun:09:00",
		},
	})

	// Make sure to delete when the function returns.
	defer cleanupAndAssert(t, terraformOptions)

	// Run the Terraform Workspace.
	terraform.InitAndApply(t, terraformOptions)

	// Initialize the Cloud-API Client
	c := client.NewClient(
		true,
		os.Getenv("APTIBLE_HOST"),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

	// Get VPC Asset ID from terraform output and use it to get the Asset from Cloud-API
	vpcId := terraform.Output(t, terraformOptions, "vpc_id")

	// check cloud api's understanding of asset
	vpcAsset, vpcAptibleErr := c.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		vpcId,
	)
	assert.Nil(t, vpcAptibleErr)
	assert.Equal(t, vpcAsset.Id, vpcId)
	assert.Equal(t, vpcAsset.Status, cac.ASSETSTATUS_DEPLOYED)

	// check aws asset state
	vpcAws, vpcAwsErr := terratest_aws.GetVpcsE(t, []*legacy_aws_sdk_ec2.Filter{
		{
			Name:   aws.String("tag:Name"),
			Values: []*string{aws.String(vpc_name)},
		},
	}, "us-east-1")
	assert.Nil(t, vpcAwsErr)
	assert.GreaterOrEqual(t, len(vpcAws), 1)
	assert.Equal(t, len(vpcAws[0].Subnets), 6)

	redisId := terraform.Output(t, terraformOptions, "redis_id")

	// check cloud api's understanding of asset
	redisAsset, redisErr := c.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		redisId,
	)
	assert.Nil(t, redisErr)
	assert.Equal(t, redisAsset.Id, redisId)
	assert.Equal(t, redisAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	assert.NotNil(t, redisAsset.Outputs)

	redis_cluster_id_base := redisAsset.GetOutputs()["elasticache_cluster_id"].Data.(string)

	cluster_details, err := GetCluster(redis_cluster_id_base + "-001")
	assert.Nil(t, err)
	assert.Equal(t, *cluster_details.CacheClusterStatus, "available")

	terraformUpdateOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"organization_id":    os.Getenv("ORGANIZATION_ID"),
			"environment_id":     os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":       os.Getenv("APTIBLE_HOST"),
			"node_name":          "test-elasticache-node",
			"vpc_name":           vpc_name,
			"description":        "TestingModule",
			"snapshot_window":    "02:00-05:00",
			"maintenance_window": "sun:06:00-sun:10:00",
		},
	})

	// Run the Terraform Workspace.
	terraform.InitAndApply(t, terraformUpdateOptions)

	cluster_update_details, err := GetCluster(redis_cluster_id_base + "-001")
	assert.Nil(t, err)
	assert.Equal(t, *cluster_update_details.CacheClusterStatus, "available")

	assert.Equal(t, *cluster_update_details.PreferredMaintenanceWindow, "sun:06:00-sun:10:00")
	assert.Equal(t, *cluster_update_details.SnapshotWindow, "02:00-05:00")
}

func GetCluster(cluster_id string) (*types.CacheCluster, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	c := elasticache.NewFromConfig(cfg)
	clusters, err := c.DescribeCacheClusters(ctx, &elasticache.DescribeCacheClustersInput{CacheClusterId: aws.String(cluster_id)})
	if err != nil {
		return nil, err
	}
	if clusters == nil || len(clusters.CacheClusters) == 0 {
		return nil, fmt.Errorf("aws returned nil object or no clusters from search parameters")
	}

	return &clusters.CacheClusters[0], nil
}
