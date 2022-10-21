package test

import (
	"context"
	"os"
	"testing"

	"golang.org/x/exp/maps"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
)

const SimpleTestCaseDir = "./basic/simple"

func case1(t *testing.T, ctx context.Context, c client.CloudClient, baseVariables map[string]interface{}) {
	testCaseVars1 := map[string]interface{}{
		"add_secret": true,
		"add_vpc":    false,
	}
	maps.Copy(testCaseVars1, baseVariables)
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: SimpleTestCaseDir,
		Vars:         testCaseVars1,
	})
	runTerratestLoop(t, terraformOptions, func() {
		secretsId := terraform.Output(t, terraformOptions, "secrets_id")
		assert.NotEmpty(t, secretsId)
		asset, err := c.DescribeAsset(
			ctx,
			baseVariables["organization_id"].(string),
			baseVariables["environment_id"].(string),
			stripBraces(secretsId),
		)
		assert.Nil(t, err)
		assert.Equal(t, asset.Id, stripBraces(secretsId))
	})
}

func case2(t *testing.T, ctx context.Context, c client.CloudClient, baseVariables map[string]interface{}) {
	testCaseVars2 := map[string]interface{}{
		"add_secret": true,
		"add_vpc":    true,
	}
	maps.Copy(testCaseVars2, baseVariables)
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: SimpleTestCaseDir,
		Vars:         testCaseVars2,
	})
	runTerratestLoop(t, terraformOptions, func() {
		secretsId := terraform.Output(t, terraformOptions, "secrets_id")
		vpcId := terraform.Output(t, terraformOptions, "vpc_id")
		assert.NotEmpty(t, vpcId)
		assert.NotEmpty(t, secretsId)
		secretsAsset, secretsErr := c.DescribeAsset(
			ctx,
			baseVariables["organization_id"].(string),
			baseVariables["environment_id"].(string),
			stripBraces(secretsId),
		)
		assert.Nil(t, secretsErr)
		assert.Equal(t, secretsAsset.Id, stripBraces(secretsId))
		vpcAsset, vpcErr := c.DescribeAsset(
			ctx,
			baseVariables["organization_id"].(string),
			baseVariables["environment_id"].(string),
			stripBraces(vpcId),
		)
		assert.Nil(t, vpcErr)
		assert.Equal(t, vpcAsset.Id, stripBraces(vpcId))
	})
}

func TestTerraformSimple(t *testing.T) {
	baseVariables := map[string]interface{}{
		"organization_id": os.Getenv("ORGANIZATION_ID"),
		"environment_id":  os.Getenv("ENVIRONMENT_ID"),
		"aptible_host":    os.Getenv("APTIBLE_HOST"),
		"fqdn":            "testing",
		"domain":          "this-should-never-work.aptible-cloud-staging.com",
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: SimpleTestCaseDir,
		Vars:         baseVariables,
	})
	defer terraform.Destroy(t, terraformOptions)

	checkEnvVars(t)
	c := client.NewClient(
		true,
		baseVariables["aptible_host"].(string),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

	t.Parallel()

	// add secret only (fast asset)
	// ---- test 1
	case1(t, ctx, c, copy(baseVariables))

	// add vpc to secrets
	// ---- test 2
	case2(t, ctx, c, copy(baseVariables))
}
