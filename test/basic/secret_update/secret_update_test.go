package secret_create

import (
	"context"
	"encoding/json"
	"os"
	"testing"

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
	"secret_name":     "testing-secret",
	"secret_value":    "some-kind-of-secret-string",
}

func getAptibleSecretAndAWSSecret(t *testing.T, ctx context.Context, client client.CloudClient, secretId string) (*cac.AssetOutput, string, string, error) {
	secretAsset, err := client.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		secretId,
	)
	if err != nil {
		return nil, "", "", err
	}
	tfOptions := generateMutableTerraformOptions()
	secretArn := terraform.Output(t, tfOptions, "secret_arn")
	secretValue := terratest_aws.GetSecretValue(t, "us-east-1", secretArn)
	return secretAsset, secretValue, secretArn, err
}

func cleanupAndAssert(t *testing.T, terraformOptions *terraform.Options) {
	terraform.Destroy(t, terraformOptions)
	// test / assert all failures here
}

// generateMutableTerraformOptions - Generates a new pointer and object of mutable reference of the variable map,
// which is mutated over the course of the test suite to avoid specifying the full set
func generateMutableTerraformOptions() *terraform.Options {
	return &terraform.Options{
		TerraformDir: ".",
		Vars:         mutableTFVariables,
	}
}

func TestSecretUpdate(t *testing.T) {
	tfOptions := generateMutableTerraformOptions()
	defer cleanupAndAssert(t, tfOptions)
	terraform.InitAndApply(t, tfOptions)

	c := client.NewClient(
		true,
		os.Getenv("APTIBLE_HOST"),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

	secretId := terraform.Output(t, tfOptions, "secret_id")

	// create string secret
	secretAsset, secretValue, secretArn, err := getAptibleSecretAndAWSSecret(t, ctx, c, secretId)
	assert.Nil(t, err)
	assert.Equal(t, secretAsset.Id, secretId)
	assert.Equal(t, secretAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	assert.Equal(t, secretValue, "some-kind-of-secret-string")

	// change secret insertion to json
	secretEncodedValue, _ := json.Marshal(map[string]string{
		"test-value-1": "test1",
		"test-value-2": "test2",
	})
	mutableTFVariables["secret_name"] = "testing-secret-but-json"
	mutableTFVariables["secret_value"] = string(secretEncodedValue)
	terraform.Apply(t, generateMutableTerraformOptions())

	secretAsset, updatedSecretValue, updatedSecretArn, err := getAptibleSecretAndAWSSecret(t, ctx, c, secretId)
	assert.Nil(t, err)
	assert.Equal(t, secretAsset.Id, secretId)
	assert.Equal(t, secretAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	assert.Equal(t, updatedSecretValue, string(secretEncodedValue))
	assert.NotEqual(t, updatedSecretArn, secretArn)

	// check it reverted to original setup
	mutableTFVariables["secret_name"] = "testing-secret"
	mutableTFVariables["secret_value"] = "some-kind-of-secret-string"
	terraform.Apply(t, generateMutableTerraformOptions())

	secretAsset, secretValue, _, err = getAptibleSecretAndAWSSecret(t, ctx, c, secretId)
	assert.Nil(t, err)
	assert.Equal(t, secretAsset.Id, secretId)
	assert.Equal(t, secretAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	assert.Equal(t, secretValue, "some-kind-of-secret-string")
}
