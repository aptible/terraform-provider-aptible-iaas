package ecs_compute_update

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	legacy_aws_sdk_ec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
)

var mutableTFVariables = map[string]interface{}{
	"organization_id":   os.Getenv("ORGANIZATION_ID"),
	"environment_id":    os.Getenv("ENVIRONMENT_ID"),
	"aptible_host":      os.Getenv("APTIBLE_HOST"),
	"compute_name":      "ecs-compute-test",
	"container_command": []string{"nginx", "-g", "daemon off;"},
	"container_image":   "nginx",
	"container_port":    80,
	"vpc_name":          "testecs-compute-vpc",
	"is_ecr_image":      false,
}

func cleanupAndAssert(t *testing.T, terraformOptions *terraform.Options) {
	terraform.Destroy(t, terraformOptions)

	// test / assert all failures here
}

func getAptibleAndAWSVPCs(t *testing.T, ctx context.Context, client client.CloudClient, vpcId, vpcName string) (*cac.AssetOutput, []*terratest_aws.Vpc, error) {
	vpcAsset, err := client.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		vpcId,
	)
	if err != nil {
		return nil, nil, err
	}

	vpcAws, err := terratest_aws.GetVpcsE(t, []*legacy_aws_sdk_ec2.Filter{
		{
			Name:   aws.String("tag:Name"),
			Values: []*string{aws.String(vpcName)},
		},
	}, "us-east-1")
	if err != nil {
		return nil, nil, err
	}

	return vpcAsset, vpcAws, nil
}

func getAptibleAndAWSECSServiceAndCluster(t *testing.T, ctx context.Context, client client.CloudClient, ecsComputeId, ecsClusterName, ecsServiceName string) (*cac.AssetOutput, *ecs.Cluster, *ecs.Service, error) {
	ecsComputeAsset, err := client.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		ecsComputeId,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	ecsClusterAws := terratest_aws.GetEcsCluster(t, "us-east-1", ecsClusterName)
	ecsServiceAws := terratest_aws.GetEcsService(t, "us-east-1", ecsClusterName, ecsServiceName)
	return ecsComputeAsset, ecsClusterAws, ecsServiceAws, nil
}

func assertCommonValues(t *testing.T, vpcId, ecsComputeId string, vpcAsset *cac.AssetOutput, vpcAws []*terratest_aws.Vpc, ecsComputeAsset *cac.AssetOutput, ecsClusterAws *ecs.Cluster, ecsServiceAws *ecs.Service) {
	// WARNING - THIS MUST BE A BLIND HELPER TO REDUCE LOC WITH ASSERTING FROM INPUT VALS PROVIDED
	assert.Equal(t, vpcAsset.Id, vpcId)
	assert.Equal(t, vpcAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	assert.GreaterOrEqual(t, len(vpcAws), 1)
	assert.Equal(t, len(vpcAws[0].Subnets), 6)
	assert.Equal(t, vpcAws[0].Tags["aptible_asset_id"], vpcId)
	assert.Equal(t, ecsComputeAsset.Id, ecsComputeId)
	assert.Equal(t, ecsComputeAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	assert.NotNil(t, ecsComputeAsset.Outputs)
	assert.Equal(t, *ecsClusterAws.Status, "ACTIVE")
	assert.Equal(t, *ecsServiceAws.Status, "ACTIVE")
}

// generateMutableTerraformOptions - Generates a new pointer and object of mutable reference of the variable map,
// which is mutated over the course of the test suite to avoid specifying the full set
func generateMutableTerraformOptions() *terraform.Options {
	return &terraform.Options{
		TerraformDir: ".",
		Vars:         mutableTFVariables,
		PlanFilePath: "out.plan",
	}
}

func TestECSComputeUpdate(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, generateMutableTerraformOptions())
	//defer cleanupAndAssert(t, generateMutableTerraformOptions())
	terraform.InitAndPlan(t, generateMutableTerraformOptions())
	terraform.Apply(t, generateMutableTerraformOptions())

	c := client.NewClient(
		true,
		os.Getenv("APTIBLE_HOST"),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

	vpcId := terraform.Output(t, terraformOptions, "vpc_id")
	vpcAsset, vpcAws, err := getAptibleAndAWSVPCs(t, ctx, c, vpcId, "testecs-compute-vpc")
	assert.Nil(t, err)

	ecsComputeId := terraform.Output(t, terraformOptions, "ecs_compute_id")
	ecsComputeAsset, ecsClusterAws, ecsServiceAws, err := getAptibleAndAWSECSServiceAndCluster(t, ctx, c, ecsComputeId, "ecs-compute-test-compute-cluster", "ecs-compute-test")
	assert.Nil(t, err)
	assertCommonValues(t, vpcId, ecsComputeId, vpcAsset, vpcAws, ecsComputeAsset, ecsClusterAws, ecsServiceAws)

	type atomicChange struct {
		key      string
		updated  interface{}
		original interface{}
	}
	for _, changeSet := range [][]atomicChange{
		// change container name
		{{
			key:      "compute_name",
			updated:  "ecs-compute-uname-only-test",
			original: "ecs-compute-test",
		}},
		// change image to nginx => httpd and then back!
		{{
			key:      "container_command",
			updated:  []string{"httpd", "-D", "FOREGROUND"},
			original: []string{"nginx", "-g", "daemon off;"},
		}, {
			key:      "container_image",
			updated:  "httpd",
			original: "nginx",
		}},
		// change image to nginx => httpd and then back!
		{{
			key:      "container_command",
			updated:  []string{"nginx", "-g", "daemon off;"},
			original: []string{"nginx", "-g", "daemon off;"},
		}, {
			key:      "container_port",
			updated:  81,
			original: 80,
		}},
		// change image to is_ecr_image to true
		{{
			key:      "container_image",
			updated:  "public.ecr.aws/nginx/nginx:1-alpine-perl",
			original: "nginx",
		}, {
			key:      "is_ecr_image",
			updated:  true,
			original: false,
		}},
	} {
		for _, change := range changeSet {
			mutableTFVariables[change.key] = change.updated
		}
		terraform.Plan(t, generateMutableTerraformOptions())
		terraform.Apply(t, generateMutableTerraformOptions())
		updatedECSComputeId := terraform.Output(t, terraformOptions, "ecs_compute_id")
		updatedECSComputeAsset, updatedECSClusterAws, updatedECSServiceAws, err := getAptibleAndAWSECSServiceAndCluster(
			t,
			ctx,
			c,
			updatedECSComputeId,
			fmt.Sprintf("%s-compute-cluster", mutableTFVariables["compute_name"].(string)),
			mutableTFVariables["compute_name"].(string),
		)
		assert.Nil(t, err)
		assertCommonValues(t, vpcId, updatedECSComputeId, vpcAsset, vpcAws, updatedECSComputeAsset, updatedECSClusterAws, updatedECSServiceAws)
		for _, change := range changeSet {
			mutableTFVariables[change.key] = change.original
		}
		terraform.Plan(t, generateMutableTerraformOptions())
		terraform.Apply(t, generateMutableTerraformOptions())
	}
}
