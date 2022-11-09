package simple

import (
	"context"
	"os"
	"testing"

	"golang.org/x/exp/maps"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	"github.com/aptible/terraform-provider-aptible-iaas/test"
)

const TestCaseDir = "."

func case1(t *testing.T, ctx context.Context, c client.CloudClient, baseVariables map[string]interface{}) {
	testCaseVars1 := map[string]interface{}{}
	maps.Copy(testCaseVars1, baseVariables)
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: TestCaseDir,
		Vars:         testCaseVars1,
	})
	test.RunTerratestLoop(t, terraformOptions, func() {
		secretsId := terraform.Output(t, terraformOptions, "secrets_id")
		assert.NotEmpty(t, secretsId)
		asset, err := c.DescribeAsset(
			ctx,
			baseVariables["organization_id"].(string),
			baseVariables["environment_id"].(string),
			test.StripBraces(secretsId),
		)
		assert.Nil(t, err)
		assert.Equal(t, asset.Id, test.StripBraces(secretsId))
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
		TerraformDir: TestCaseDir,
		Vars:         baseVariables,
	})
	defer terraform.Destroy(t, terraformOptions)

	test.CheckEnvVars(t)
	c := client.NewClient(
		true,
		baseVariables["aptible_host"].(string),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

	t.Parallel()

	// add everything (fast asset)
	// ---- test 1
	case1(t, ctx, c, test.Copy(baseVariables))
}
