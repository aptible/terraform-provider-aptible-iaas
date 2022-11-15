package postgres

import (
	"context"
	"os"
	"testing"

	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	"github.com/aptible/terraform-provider-aptible-iaas/test/basic/rds_create"
)

func cleanupAndAssert(t *testing.T, terraformOptions *terraform.Options) {
	terraform.Destroy(t, terraformOptions)

	// test / assert all failures here
}

func TestRDSCreatePostgres(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"organization_id":         os.Getenv("ORGANIZATION_ID"),
			"environment_id":          os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":            os.Getenv("APTIBLE_HOST"),
			"database_name":           "test-create-postgres-14",
			"database_engine_version": "14",
			"vpc_name":                "rds-create-vpc-pg-14",
		},
	})
	defer cleanupAndAssert(t, terraformOptions)

	err := rds_create.CheckOrRequestVPCLimit()
	if assert.Nil(t, err) {
		t.Fail()
	}
	terraform.InitAndApply(t, terraformOptions)

	c := client.NewClient(
		true,
		os.Getenv("APTIBLE_HOST"),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

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
