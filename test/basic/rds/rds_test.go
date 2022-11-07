package rds

import (
	"context"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
)

func TestRDS(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars: map[string]interface{}{
			"organization_id": os.Getenv("ORGANIZATION_ID"),
			"environment_id":  os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":    os.Getenv("APTIBLE_HOST"),
			"database_name":   "testrds",
			"vpc_name":        "testrds",
		},
	})
	defer terraform.Destroy(t, terraformOptions)

	c := client.NewClient(
		true,
		os.Getenv("APTIBLE_HOST"),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

	vpcId := terraform.Output(t, terraformOptions, "vpc_id")
	// check cloud api's understanding of asset
	vpcAsset, vpcErr := c.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		vpcId[1:len(vpcId)-1],
	)
	assert.Nil(t, vpcErr)
	assert.Equal(t, vpcAsset.Id, vpcId)
	assert.Equal(t, vpcAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	// check aws asset state

	rdsId := terraform.Output(t, terraformOptions, "rds_id")
	// check cloud api's understanding of asset
	rdsAsset, rdsErr := c.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		rdsId[1:len(rdsId)-1],
	)
	assert.Nil(t, rdsErr)
	assert.Equal(t, rdsAsset.Id, rdsId)
	assert.Equal(t, rdsAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	// check aws asset state

}
