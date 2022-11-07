package ecs_compute

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	legacy_aws_sdk_ec2 "github.com/aws/aws-sdk-go/service/ec2"
	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
)

func cleanupAndAssert(t *testing.T, terraformOptions *terraform.Options) {
	terraform.Destroy(t, terraformOptions)

	// test / assert all failures here
}

func TestECSCompute(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"organization_id":   os.Getenv("ORGANIZATION_ID"),
			"environment_id":    os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":      os.Getenv("APTIBLE_HOST"),
			"compute_name":      "ecs-compute-test",
			"container_command": []string{"nginx", "-g", "daemon off;"},
			"container_image":   "nginx",
			"container_port":    80,
			"vpc_name":          "testecs-compute-vpc",
		},
	})
	defer cleanupAndAssert(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	c := client.NewClient(
		true,
		os.Getenv("APTIBLE_HOST"),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

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
			Values: []*string{aws.String("testecs-compute-vpc")},
		},
	}, "us-east-1")
	assert.Nil(t, vpcAwsErr)
	assert.GreaterOrEqual(t, len(vpcAws), 1)
	assert.Equal(t, len(vpcAws[0].Subnets), 6)

	ecsComputeId := terraform.Output(t, terraformOptions, "ecs_compute_id")
	// check cloud api's understanding of asset
	ecsComputeAsset, ecsComputeErr := c.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		ecsComputeId,
	)
	assert.Nil(t, ecsComputeErr)
	assert.Equal(t, ecsComputeAsset.Id, ecsComputeId)
	assert.Equal(t, ecsComputeAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	assert.NotNil(t, ecsComputeAsset.Outputs)

	// check aws asset state
	ecsClusterAws, ecsClusterAwsErr := terratest_aws.GetEcsClusterE(t, "us-east-1", "ecs-compute-test-compute-cluster")
	assert.Nil(t, ecsClusterAwsErr)
	assert.Equal(t, *ecsClusterAws.Status, "ACTIVE")
	ecsServiceAws, ecsServiceAwserr := terratest_aws.GetEcsServiceE(t, "us-east-1", "ecs-compute-test-compute-cluster", "ecs-compute-test")
	assert.Nil(t, ecsServiceAwserr)
	assert.Equal(t, *ecsServiceAws.Status, "ACTIVE")
}
