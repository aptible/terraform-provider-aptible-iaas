package aurora

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

func TestRDSCreatePostgres(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"organization_id": os.Getenv("ORGANIZATION_ID"),
			"environment_id":  os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":    os.Getenv("APTIBLE_HOST"),
			"database_name":   "testrds-aurora",
			"vpc_name":        "testrds-vpc",
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
			Values: []*string{aws.String("testrds-vpc")},
		},
	}, "us-east-1")
	assert.Nil(t, vpcAwsErr)
	assert.GreaterOrEqual(t, len(vpcAws), 1)
	assert.Equal(t, len(vpcAws[0].Subnets), 6)

	rdsId := terraform.Output(t, terraformOptions, "rds_id")
	rdsInstanceId := terraform.Output(t, terraformOptions, "rds_db_identifier")
	// check cloud api's understanding of asset
	rdsAsset, rdsErr := c.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		rdsId,
	)
	assert.Nil(t, rdsErr)
	assert.Equal(t, rdsAsset.Id, rdsId)
	assert.Equal(t, rdsAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	assert.NotNil(t, rdsAsset.Outputs)
	assert.Equal(t, rdsAsset.GetOutputs()["db_identifier"].Data.(string), rdsInstanceId)

	// check aws asset state
	rdsAws, rdsAwsErr := terratest_aws.GetRdsInstanceDetailsE(t, rdsInstanceId, "us-east-1")
	assert.Nil(t, rdsAwsErr)
	assert.Equal(t, *rdsAws.DBInstanceStatus, "available")
}
