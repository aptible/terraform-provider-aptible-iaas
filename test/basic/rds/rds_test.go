package rds

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
)

func TestRDS(t *testing.T) {

	vpc_name := "testrds"
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars: map[string]interface{}{
			"organization_id": os.Getenv("ORGANIZATION_ID"),
			"environment_id":  os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":    os.Getenv("APTIBLE_HOST"),
			"database_name":   "testrds",
			"vpc_name":        vpc_name,
		},
	})
	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	c := client.NewClient(
		true,
		os.Getenv("APTIBLE_HOST"),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}
	client := ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		o.Region = "us-east-1"
	})

	vpcId := terraform.Output(t, terraformOptions, "vpc_id")
	// check cloud api's understanding of asset
	vpcAsset, vpcAptibleErr := c.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		vpcId[1:len(vpcId)-1],
	)
	assert.Nil(t, vpcAptibleErr)
	assert.Equal(t, vpcAsset.Id, vpcId)
	assert.Equal(t, vpcAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	// check aws asset state

	vpcAws, vpcAwsErr := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{Filters: []types.Filter{
		{Name: aws.String("tag:Name"),
			Values: []string{vpc_name}},
	}})
	assert.Nil(t, vpcAwsErr)
	assert.GreaterOrEqual(t, len(vpcAws.Vpcs), 1)
	assert.Equal(t, vpcAws.Vpcs[0].State, types.VpcStateAvailable)

	// rdsId := terraform.Output(t, terraformOptions, "rds_id")
	// // check cloud api's understanding of asset
	// rdsAsset, rdsErr := c.DescribeAsset(
	// 	ctx,
	// 	os.Getenv("ORGANIZATION_ID"),
	// 	os.Getenv("ENVIRONMENT_ID"),
	// 	rdsId[1:len(rdsId)-1],
	// )
	// assert.Nil(t, rdsErr)
	// assert.Equal(t, rdsAsset.Id, rdsId)
	// assert.Equal(t, rdsAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	// check aws asset state

}
