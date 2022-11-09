package vpc_update

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

var mutableTFVariables = map[string]interface{}{
	"organization_id": os.Getenv("ORGANIZATION_ID"),
	"environment_id":  os.Getenv("ENVIRONMENT_ID"),
	"aptible_host":    os.Getenv("APTIBLE_HOST"),
	"vpc_name":        "test-vpc",
}

func assertCommonValues(t *testing.T, vpcId string, vpcAsset *cac.AssetOutput, vpcAws []*terratest_aws.Vpc) {
	assert.Equal(t, vpcAsset.Id, vpcId)
	assert.Equal(t, vpcAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	assert.GreaterOrEqual(t, len(vpcAws), 1)
	assert.Equal(t, len(vpcAws[0].Subnets), 6)
}

// generateMutableTerraformOptions - Generates a new pointer and object of mutable reference of the variable map,
// which is mutated over the course of the test suite to avoid specifying the full set
func generateMutableTerraformOptions() *terraform.Options {
	return &terraform.Options{
		TerraformDir: ".",
		Vars:         mutableTFVariables,
	}
}

func reachabilityAnalysisAssertions(t *testing.T) {
	privInstanceENI := terraform.Output(t, generateMutableTerraformOptions(), "test_instance_private_eni")
	pubInstanceENI := terraform.Output(t, generateMutableTerraformOptions(), "test_instance_public_eni")
	analysisId := terraform.Output(t, generateMutableTerraformOptions(), "analysis_id")
	insightsId := terraform.Output(t, generateMutableTerraformOptions(), "insights_id")

	ec2 := terratest_aws.NewEc2Client(t, "us-east-1")
	networkInsights, networkInsightsErr := ec2.DescribeNetworkInsightsPaths(&legacy_aws_sdk_ec2.DescribeNetworkInsightsPathsInput{
		NetworkInsightsPathIds: []*string{aws.String(insightsId)},
	})
	assert.Nil(t, networkInsightsErr)
	assert.Equal(t, len(networkInsights.NetworkInsightsPaths), 1)
	assert.Equal(t, *networkInsights.NetworkInsightsPaths[0].Destination, pubInstanceENI)
	assert.Equal(t, *networkInsights.NetworkInsightsPaths[0].Source, privInstanceENI)

	networkAnalysis, networkAnalysisErr := ec2.DescribeNetworkInsightsAnalyses(
		&legacy_aws_sdk_ec2.DescribeNetworkInsightsAnalysesInput{
			NetworkInsightsAnalysisIds: []*string{aws.String(analysisId)},
		})
	assert.Nil(t, networkAnalysisErr)
	assert.Equal(t, len(networkAnalysis.NetworkInsightsAnalyses), 1)
	assert.Equal(t, *networkAnalysis.NetworkInsightsAnalyses[0].Status, "succeeded")
}

func TestVPCUpdate(t *testing.T) {
	defer terraform.Destroy(t, generateMutableTerraformOptions())
	terraform.InitAndApply(t, generateMutableTerraformOptions())

	c := client.NewClient(
		true,
		os.Getenv("APTIBLE_HOST"),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

	vpcId := terraform.Output(t, generateMutableTerraformOptions(), "vpc_id")
	// check cloud api's understanding of asset
	vpcAsset, err := c.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		vpcId,
	)
	assert.Nil(t, err)
	// check aws asset state
	vpcAws, err := terratest_aws.GetVpcsE(t, []*legacy_aws_sdk_ec2.Filter{
		{
			Name:   aws.String("tag:Name"),
			Values: []*string{aws.String("test-vpc")},
		},
	}, "us-east-1")
	assert.Nil(t, err)
	assertCommonValues(t, vpcId, vpcAsset, vpcAws)
	assert.Equal(t, vpcAws[0].Name, "test-vpc")
	reachabilityAnalysisAssertions(t)

	mutableTFVariables["vpc_name"] = "test-vpc-updated"
	terraform.Apply(t, generateMutableTerraformOptions())

	updatedVpcId := terraform.Output(t, generateMutableTerraformOptions(), "vpc_id")
	// check cloud api's understanding of asset
	updatedVpcAsset, err := c.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		updatedVpcId,
	)
	assert.Nil(t, err)
	// check aws asset state
	updatedVpcAws, err := terratest_aws.GetVpcsE(t, []*legacy_aws_sdk_ec2.Filter{
		{
			Name:   aws.String("tag:Name"),
			Values: []*string{aws.String("test-vpc-updated")},
		},
	}, "us-east-1")
	assert.Nil(t, err)
	assertCommonValues(t, updatedVpcId, updatedVpcAsset, updatedVpcAws)
	assert.Equal(t, updatedVpcAws[0].Name, "test-vpc-updated")
	reachabilityAnalysisAssertions(t)

	mutableTFVariables["vpc_name"] = "test-vpc"
	terraform.Apply(t, generateMutableTerraformOptions())
}
