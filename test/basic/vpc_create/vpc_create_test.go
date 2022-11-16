package vpc_create

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	legacy_aws_sdk_ec2 "github.com/aws/aws-sdk-go/service/ec2"
	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	"github.com/aptible/terraform-provider-aptible-iaas/test/utils"
)

func cleanupAndAssert(ctx context.Context, t *testing.T, terraformOptions *terraform.Options, environmentId, assetId string) {
	terraform.Destroy(t, terraformOptions)

	resources, err := utils.GetTaggedResources(ctx, environmentId, assetId)
	fmt.Printf("Resource response is %v", resources)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resources))
}

func TestVPCCreate(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"organization_id": os.Getenv("ORGANIZATION_ID"),
			"environment_id":  os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":    os.Getenv("APTIBLE_HOST"),
			"vpc_name":        "test-vpc",
		},
	})

	// This is the destroy of last resort. By the time it runs everything should have
	// already been destroyed, but in case of failure we do one more pass.
	defer terraform.Destroy(t, terraformOptions)

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

	// Runs cleanup and assertions. This needs the asset id which is why we launch it down here.
	// This will run before the defers above, so this destroy should happen before the destroy
	// of last resort.
	defer cleanupAndAssert(ctx, t, terraformOptions, os.Getenv("ENVIRONMENT_ID"), vpcAsset.Id)

	// check aws asset state
	vpcAws, vpcAwsErr := terratest_aws.GetVpcsE(t, []*legacy_aws_sdk_ec2.Filter{
		{
			Name:   aws.String("tag:Name"),
			Values: []*string{aws.String("test-vpc")},
		},
	}, "us-east-1")
	assert.Nil(t, vpcAwsErr)
	assert.GreaterOrEqual(t, len(vpcAws), 1)
	assert.Equal(t, len(vpcAws[0].Subnets), 6)
	assert.Equal(t, vpcAws[0].Name, "test-vpc")

	privInstanceENI := terraform.Output(t, terraformOptions, "test_instance_private_eni")
	pubInstanceENI := terraform.Output(t, terraformOptions, "test_instance_public_eni")
	analysisId := terraform.Output(t, terraformOptions, "analysis_id")
	insightsId := terraform.Output(t, terraformOptions, "insights_id")

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

	tagged_resources, err := utils.GetTaggedResources(ctx, os.Getenv("ENVIRONMENT_ID"), vpcAsset.Id)
	assert.Nil(t, err)
	assert.Greater(t, len(tagged_resources), 0)
}
