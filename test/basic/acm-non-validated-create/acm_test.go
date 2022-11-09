package acm

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	legacy_aws_sdk_acm "github.com/aws/aws-sdk-go/service/acm"
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

func TestACMNonValidated(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars: map[string]interface{}{
			"organization_id": os.Getenv("ORGANIZATION_ID"),
			"environment_id":  os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":    os.Getenv("APTIBLE_HOST"),
			"domain":          "aptible-cloud-staging.com",
			"subdomain":       "never-should-be-valid",
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

	certId := terraform.Output(t, terraformOptions, "cert_id")

	// check cloud api's understanding of asset
	certAsset, certAptibleErr := c.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		certId,
	)
	assert.Nil(t, certAptibleErr)
	assert.Equal(t, certAsset.Id, certId)
	assert.Equal(t, certAsset.Status, cac.ASSETSTATUS_DEPLOYED)

	// check aws asset state
	certArn := terraform.Output(t, terraformOptions, "cert_arn")
	certAccountId := terraform.Output(t, terraformOptions, "aptible_aws_account_id")
	session, sessionErr := terratest_aws.CreateAwsSessionFromRole("us-east-1", fmt.Sprintf("arn:aws:iam::%s:role/OrganizationAccountAccessRole", certAccountId))
	if sessionErr != nil {
		fmt.Println(sessionErr.Error())
		os.Exit(1)
	}

	acmClient := legacy_aws_sdk_acm.New(session)
	certAws, certAwsErr := acmClient.DescribeCertificate(&legacy_aws_sdk_acm.DescribeCertificateInput{
		CertificateArn: aws.String(certArn),
	})
	assert.Nil(t, certAwsErr)
	assert.Equal(t, legacy_aws_sdk_acm.CertificateStatusPendingValidation, *certAws.Certificate.Status)

	domainValidation := terraform.OutputListOfObjects(t, terraformOptions, "domain_validation_records")
	assert.Equal(t, 1, len(domainValidation))
	assert.Equal(t, domainValidation[0]["domain_name"], "never-should-be-valid.aptible-cloud-staging.com")
	assert.Equal(t, domainValidation[0]["resource_record_type"], "CNAME")

	certFdqn := terraform.Output(t, terraformOptions, "fqdn")
	assert.Equal(t, certFdqn, "never-should-be-valid.aptible-cloud-staging.com")
}
