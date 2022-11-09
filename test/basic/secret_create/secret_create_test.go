package secret_create

import (
	"context"
	"os"
	"testing"

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

func TestSecretCreate(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"organization_id": os.Getenv("ORGANIZATION_ID"),
			"environment_id":  os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":    os.Getenv("APTIBLE_HOST"),
			"secret_name":     "testing-secret",
			"secret_string":   "some-kind-of-secret-string",
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

	secretId := terraform.Output(t, terraformOptions, "secret_id")

	// check cloud api's understanding of asset
	secretAsset, secretAptibleErr := c.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		secretId,
	)
	assert.Nil(t, secretAptibleErr)
	assert.Equal(t, secretAsset.Id, secretId)
	assert.Equal(t, secretAsset.Status, cac.ASSETSTATUS_DEPLOYED)

	secretArn := terraform.Output(t, terraformOptions, "secret_arn")

	// check aws asset state
	secretValue := terratest_aws.GetSecretValue(t, "us-east-1", secretArn)
	assert.Equal(t, secretValue, "some-kind-of-secret-string")
}
