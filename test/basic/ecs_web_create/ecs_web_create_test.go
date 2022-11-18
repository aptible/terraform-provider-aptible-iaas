package ecs_compute_update

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	legacy_aws_sdk_ec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
)

func checkSetup() {
	for _, envVarKey := range []string{
		"AWS_DNS_ROLE",
	} {
		_, envVar := os.LookupEnv(envVarKey)
		if !envVar {
			fmt.Printf("%s environment variable not set\n", envVarKey)
			os.Exit(1)
		}
	}
}

func cleanupAndAssert(t *testing.T, terraformOptions *terraform.Options) {
	terraform.Destroy(t, terraformOptions)

	// test / assert all failures here
}

func getAptibleAndAWSVPCs(t *testing.T, ctx context.Context, client client.CloudClient, vpcId, vpcName string) (*cac.AssetOutput, []*terratest_aws.Vpc, error) {
	vpcAsset, err := client.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		vpcId,
	)
	if err != nil {
		return nil, nil, err
	}

	vpcAws, err := terratest_aws.GetVpcsE(t, []*legacy_aws_sdk_ec2.Filter{
		{
			Name:   aws.String("tag:Name"),
			Values: []*string{aws.String(vpcName)},
		},
	}, "us-east-1")
	if err != nil {
		return nil, nil, err
	}

	return vpcAsset, vpcAws, nil
}

func getAptibleAndAWSECSServiceAndCluster(t *testing.T, ctx context.Context, client client.CloudClient, ecsComputeId, ecsClusterName, ecsServiceName string) (*cac.AssetOutput, *ecs.Cluster, *ecs.Service, error) {
	ecsComputeAsset, err := client.DescribeAsset(
		ctx,
		os.Getenv("ORGANIZATION_ID"),
		os.Getenv("ENVIRONMENT_ID"),
		ecsComputeId,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	ecsClusterAws := terratest_aws.GetEcsCluster(t, "us-east-1", ecsClusterName)
	ecsServiceAws := terratest_aws.GetEcsService(t, "us-east-1", ecsClusterName, ecsServiceName)

	return ecsComputeAsset, ecsClusterAws, ecsServiceAws, nil
}

func insecureHttpClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}
	return client
}

func init() {
	checkSetup()
}

func assertCommonVpc(t *testing.T, vpcId string, vpcAsset *cac.AssetOutput, vpcAws []*terratest_aws.Vpc) {
	assert.Equal(t, vpcAsset.Id, vpcId)
	assert.Equal(t, vpcAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	assert.GreaterOrEqual(t, len(vpcAws), 1)
	assert.Equal(t, len(vpcAws[0].Subnets), 6)
	assert.Equal(t, vpcAws[0].Tags["aptible_asset_id"], vpcId)
}

func assertCommonEcs(t *testing.T, ecsWebId string, ecsWebAsset *cac.AssetOutput, ecsClusterAws *ecs.Cluster, ecsServiceAws *ecs.Service) {
	assert.Equal(t, ecsWebAsset.Id, ecsWebId)
	assert.Equal(t, ecsWebAsset.Status, cac.ASSETSTATUS_DEPLOYED)
	assert.NotNil(t, ecsWebAsset.Outputs)
	assert.Equal(t, *ecsClusterAws.Status, "ACTIVE")
	assert.Equal(t, *ecsServiceAws.Status, "ACTIVE")
}

func TestECSWebCreatePublicImage(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",

		Vars: map[string]interface{}{
			"organization_id":   os.Getenv("ORGANIZATION_ID"),
			"environment_id":    os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":      os.Getenv("APTIBLE_HOST"),
			"aws_dns_role":      os.Getenv("AWS_DNS_ROLE"),
			"ecs_name":          "ecs-pub-web-test",
			"container_command": []string{"nginx", "-g", "daemon off;"},
			"container_image":   "nginx",
			"container_port":    80,
			"container_name":    "nginx",
			"is_public":         true,
			"is_ecr_image":      false,
			"vpc_name":          "testecs-pub-img-web-vpc",
			"domain":            "aptible-cloud-staging.com",
			"subdomain":         "test-ecs-pub-integration",
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
	vpcAsset, vpcAws, vpcErr := getAptibleAndAWSVPCs(t, ctx, c, vpcId, "testecs-pub-img-web-vpc")
	if assert.NoError(t, vpcErr) {
		assertCommonVpc(t, vpcId, vpcAsset, vpcAws)
	}

	ecsWebId := terraform.Output(t, terraformOptions, "ecs_web_id")
	ecsWebAsset, ecsClusterAws, ecsServiceAws, ecsErr := getAptibleAndAWSECSServiceAndCluster(t, ctx, c, ecsWebId, "ecs-pub-web-test-web-cluster", "ecs-pub-web-test")
	if assert.NoError(t, ecsErr) {
		assertCommonEcs(t, ecsWebId, ecsWebAsset, ecsClusterAws, ecsServiceAws)
	}

	ecsLoadBalancerUrl := terraform.Output(t, terraformOptions, "loadbalancer_url")
	ecsLbGet, ecsLbGetErr := insecureHttpClient().Get(fmt.Sprintf("https://%s", ecsLoadBalancerUrl))
	if assert.NoError(t, ecsLbGetErr) {
		assert.EqualValues(t, ecsLbGet.StatusCode, 200)
	}

	ecsServiceUrl := terraform.Output(t, terraformOptions, "web_url")
	ecsUrlGet, ecsUrlGetErr := http.Get(fmt.Sprintf("https://%s", ecsServiceUrl))
	if assert.NoError(t, ecsUrlGetErr) {
		assert.EqualValues(t, ecsUrlGet.StatusCode, 200)
	}
}
