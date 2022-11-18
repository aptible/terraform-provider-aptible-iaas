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
		"DNS_AWS_ACCOUNT_ID",
		"SECRET_REGISTRY_USERNAME",
		"SECRET_REGISTRY_PASSWORD",
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

func TestECSWebPublicImageToPrivate(t *testing.T) {
	initialVariables := map[string]interface{}{
		"organization_id":   os.Getenv("ORGANIZATION_ID"),
		"environment_id":    os.Getenv("ENVIRONMENT_ID"),
		"aptible_host":      os.Getenv("APTIBLE_HOST"),
		"dns_account_id":    os.Getenv("DNS_AWS_ACCOUNT_ID"),
		"ecs_name":          "ecs-pub-web-test",
		"container_command": []string{"nginx", "-g", "daemon off;"},
		"container_image":   "nginx",
		"container_port":    80,
		"container_name":    "nginx",
		"is_public":         true,
		"is_ecr_image":      false,
		"vpc_name":          "testecs-pub-to-priv-vpc",
		"domain":            "aptible-cloud-staging.com",
		"subdomain":         "test-ecs-integration",
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars:         initialVariables,
	})

	defer cleanupAndAssert(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	c := client.NewClient(
		true,
		os.Getenv("APTIBLE_HOST"),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

	aptibleAccountId := terraform.Output(t, terraformOptions, "aptible_aws_account_id")
	aptibleAccountRole := fmt.Sprintf("arn:aws:iam::%s:role/OrganizationAccountAccessRole", aptibleAccountId)
	os.Setenv(terratest_aws.AuthAssumeRoleEnvVar, aptibleAccountRole)

	vpcId := terraform.Output(t, terraformOptions, "vpc_id")
	vpcAsset, vpcAws, vpcErr := getAptibleAndAWSVPCs(t, ctx, c, vpcId, "testecs-pub-to-priv-vpc")
	if assert.NoError(t, vpcErr) {
		assertCommonVpc(t, vpcId, vpcAsset, vpcAws)
	}

	ecsWebId := terraform.Output(t, terraformOptions, "ecs_web_id")
	ecsWebAsset, ecsClusterAws, ecsServiceAws, ecsErr := getAptibleAndAWSECSServiceAndCluster(t, ctx, c, ecsWebId, "ecs-pub-web-test-web-cluster", "ecs-pub-web-test")
	if assert.NoError(t, ecsErr) {
		assertCommonEcs(t, ecsWebId, ecsWebAsset, ecsClusterAws, ecsServiceAws)
	}
	assert.Equal(t, int64(1), *ecsServiceAws.DesiredCount)
	taskDefinition := terratest_aws.GetEcsTaskDefinition(t, "us-east-1", *ecsServiceAws.TaskDefinition)
	assert.Equal(t, "awsvpc", *taskDefinition.NetworkMode)
	assert.Equal(t, 2, len(taskDefinition.ContainerDefinitions)) // ssl proxy and service container
	var nginxFound bool
	for _, container := range taskDefinition.ContainerDefinitions {
		if *container.Name == "nginx" {
			nginxFound = true
			assert.Equal(t, "nginx", *container.Image)
			assert.Nil(t, container.RepositoryCredentials)
		}
	}
	assert.True(t, nginxFound)

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

	terraformRegistrySecretOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./registry_secrets",
		Vars: map[string]interface{}{
			"organization_id":          os.Getenv("ORGANIZATION_ID"),
			"environment_id":           os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":             os.Getenv("APTIBLE_HOST"),
			"registry_secret_name":     "registry123",
			"registry_secret_username": os.Getenv("SECRET_REGISTRY_USERNAME"),
			"registry_secret_password": os.Getenv("SECRET_REGISTRY_PASSWORD"),
		},
	})
	terraform.InitAndApply(t, terraformRegistrySecretOptions)
	defer terraform.Destroy(t, terraformRegistrySecretOptions)
	registrySecretArn := terraform.Output(t, terraformRegistrySecretOptions, "registry_secret_arn")

	terraformUpdateOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars: map[string]interface{}{
			"organization_id":          os.Getenv("ORGANIZATION_ID"),
			"environment_id":           os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":             os.Getenv("APTIBLE_HOST"),
			"dns_account_id":           os.Getenv("DNS_AWS_ACCOUNT_ID"),
			"ecs_name":                 "ecs-pub-web-test",
			"container_command":        []string{"nginx", "-g", "daemon off;"},
			"container_image":          "ghcr.io/aptible/docker-hello-world-private:main",
			"container_port":           80,
			"container_name":           "nginx",
			"is_public":                true,
			"is_ecr_image":             false,
			"vpc_name":                 "testecs-pub-to-priv-vpc",
			"domain":                   "aptible-cloud-staging.com",
			"subdomain":                "test-ecs-integration",
			"registry_credentials_arn": registrySecretArn,
		},
	})
	terraform.InitAndApply(t, terraformUpdateOptions)

	vpUpdateId := terraform.Output(t, terraformUpdateOptions, "vpc_id")
	assert.Equal(t, vpUpdateId, vpcId)
	vpcUpdateAsset, vpcUpdateAws, vpcUpdateErr := getAptibleAndAWSVPCs(t, ctx, c, vpUpdateId, "testecs-pub-to-priv-vpc")
	if assert.NoError(t, vpcUpdateErr) {
		assertCommonVpc(t, vpUpdateId, vpcUpdateAsset, vpcUpdateAws)
	}

	ecsUpdateWebId := terraform.Output(t, terraformUpdateOptions, "ecs_web_id")
	assert.Equal(t, ecsUpdateWebId, ecsWebId)
	ecsUpdateWebAsset, ecsUpdateClusterAws, ecsUpdateServiceAws, ecsUpdateErr := getAptibleAndAWSECSServiceAndCluster(t, ctx, c, ecsUpdateWebId, "ecs-pub-web-test-web-cluster", "ecs-pub-web-test")
	if assert.NoError(t, ecsUpdateErr) {
		assertCommonEcs(t, ecsUpdateWebId, ecsUpdateWebAsset, ecsUpdateClusterAws, ecsUpdateServiceAws)
	}
	assert.Equal(t, int64(1), *ecsUpdateServiceAws.DesiredCount)
	taskDefinitionUpdate := terratest_aws.GetEcsTaskDefinition(t, "us-east-1", *ecsUpdateServiceAws.TaskDefinition)
	assert.Equal(t, "awsvpc", *taskDefinitionUpdate.NetworkMode)
	assert.Equal(t, 2, len(taskDefinitionUpdate.ContainerDefinitions)) // ssl proxy and service container
	var nginxUpdateFound bool
	for _, container := range taskDefinitionUpdate.ContainerDefinitions {
		nginxUpdateFound = true
		if *container.Name == "nginx" {
			assert.Equal(t, "ghcr.io/aptible/docker-hello-world-private:main", *container.Image)
			assert.Equal(t, registrySecretArn, *container.RepositoryCredentials.CredentialsParameter)
		}
	}
	assert.True(t, nginxUpdateFound)

	ecsUpdateLoadBalancerUrl := terraform.Output(t, terraformUpdateOptions, "loadbalancer_url")
	ecsUpdateLbGet, ecsUpdateLbGetErr := insecureHttpClient().Get(fmt.Sprintf("https://%s", ecsUpdateLoadBalancerUrl))
	if assert.NoError(t, ecsUpdateLbGetErr) {
		assert.EqualValues(t, ecsUpdateLbGet.StatusCode, 200)
	}

	ecsUpdateServiceUrl := terraform.Output(t, terraformUpdateOptions, "web_url")
	ecsUpdateUrlGet, ecsUpdateUrlGetErr := http.Get(fmt.Sprintf("https://%s", ecsUpdateServiceUrl))
	if assert.NoError(t, ecsUpdateUrlGetErr) {
		assert.EqualValues(t, ecsUpdateUrlGet.StatusCode, 200)
	}

	terraformRevertOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars:         initialVariables,
	})
	terraform.InitAndApply(t, terraformRevertOptions)

	vpcRevertId := terraform.Output(t, terraformRevertOptions, "vpc_id")
	assert.Equal(t, vpcRevertId, vpcId)
	vpcRevertAsset, vpcRevertAws, vpcRevertErr := getAptibleAndAWSVPCs(t, ctx, c, vpcRevertId, "testecs-pub-to-priv-vpc")
	if assert.NoError(t, vpcRevertErr) {
		assertCommonVpc(t, vpcRevertId, vpcRevertAsset, vpcRevertAws)
	}

	ecsRevertWebId := terraform.Output(t, terraformRevertOptions, "ecs_web_id")
	assert.Equal(t, ecsRevertWebId, ecsWebId)
	ecsRevertWebAsset, ecsRevertClusterAws, ecsRevertServiceAws, ecsRevertErr := getAptibleAndAWSECSServiceAndCluster(t, ctx, c, ecsRevertWebId, "ecs-pub-web-test-web-cluster", "ecs-pub-web-test")
	if assert.NoError(t, ecsRevertErr) {
		assertCommonEcs(t, ecsRevertWebId, ecsRevertWebAsset, ecsRevertClusterAws, ecsRevertServiceAws)
	}
	assert.Equal(t, int64(1), *ecsRevertServiceAws.DesiredCount)
	taskDefinitionRevert := terratest_aws.GetEcsTaskDefinition(t, "us-east-1", *ecsRevertServiceAws.TaskDefinition)
	assert.Equal(t, "awsvpc", *taskDefinitionRevert.NetworkMode)
	assert.Equal(t, 2, len(taskDefinitionRevert.ContainerDefinitions)) // ssl proxy and service container
	var nginxRevertFound bool
	for _, container := range taskDefinitionRevert.ContainerDefinitions {
		nginxRevertFound = true
		if *container.Name == "nginx" {
			assert.Equal(t, "nginx", *container.Image)
			assert.Nil(t, container.RepositoryCredentials)
		}
	}
	assert.True(t, nginxRevertFound)

	ecsRevertLoadBalancerUrl := terraform.Output(t, terraformRevertOptions, "loadbalancer_url")
	ecsRevertLbGet, ecsRevertLbGetErr := insecureHttpClient().Get(fmt.Sprintf("https://%s", ecsRevertLoadBalancerUrl))
	if assert.NoError(t, ecsRevertLbGetErr) {
		assert.EqualValues(t, ecsRevertLbGet.StatusCode, 200)
	}

	ecsRevertServiceUrl := terraform.Output(t, terraformRevertOptions, "web_url")
	ecsRevertUrlGet, ecsRevertUrlGetErr := http.Get(fmt.Sprintf("https://%s", ecsRevertServiceUrl))
	if assert.NoError(t, ecsRevertUrlGetErr) {
		assert.EqualValues(t, ecsRevertUrlGet.StatusCode, 200)
	}
}

func TestECSWebChangeSecrets(t *testing.T) {
	initialVariables := map[string]interface{}{
		"organization_id":   os.Getenv("ORGANIZATION_ID"),
		"environment_id":    os.Getenv("ENVIRONMENT_ID"),
		"aptible_host":      os.Getenv("APTIBLE_HOST"),
		"dns_account_id":    os.Getenv("DNS_AWS_ACCOUNT_ID"),
		"ecs_name":          "ecs-pub-web-test-secrets",
		"container_command": []string{"nginx", "-g", "daemon off;"},
		"container_image":   "nginx",
		"container_port":    80,
		"container_name":    "nginx",
		"is_public":         true,
		"is_ecr_image":      false,
		"vpc_name":          "testecs-secrets-vpc",
		"domain":            "aptible-cloud-staging.com",
		"subdomain":         "test-ecs-integration",
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars:         initialVariables,
	})

	defer cleanupAndAssert(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	c := client.NewClient(
		true,
		os.Getenv("APTIBLE_HOST"),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

	aptibleAccountId := terraform.Output(t, terraformOptions, "aptible_aws_account_id")
	aptibleAccountRole := fmt.Sprintf("arn:aws:iam::%s:role/OrganizationAccountAccessRole", aptibleAccountId)
	os.Setenv(terratest_aws.AuthAssumeRoleEnvVar, aptibleAccountRole)

	vpcId := terraform.Output(t, terraformOptions, "vpc_id")
	vpcAsset, vpcAws, vpcErr := getAptibleAndAWSVPCs(t, ctx, c, vpcId, "testecs-secrets-vpc")
	if assert.NoError(t, vpcErr) {
		assertCommonVpc(t, vpcId, vpcAsset, vpcAws)
	}

	ecsWebId := terraform.Output(t, terraformOptions, "ecs_web_id")
	ecsWebAsset, ecsClusterAws, ecsServiceAws, ecsErr := getAptibleAndAWSECSServiceAndCluster(t, ctx, c, ecsWebId, "ecs-pub-web-test-secrets-web-cluster", "ecs-pub-web-test-secrets")
	if assert.NoError(t, ecsErr) {
		assertCommonEcs(t, ecsWebId, ecsWebAsset, ecsClusterAws, ecsServiceAws)
	}
	assert.Equal(t, int64(1), *ecsServiceAws.DesiredCount)
	taskDefinition := terratest_aws.GetEcsTaskDefinition(t, "us-east-1", *ecsServiceAws.TaskDefinition)
	assert.Equal(t, "awsvpc", *taskDefinition.NetworkMode)
	assert.Equal(t, 2, len(taskDefinition.ContainerDefinitions)) // ssl proxy and service container
	var nginxFound bool
	for _, container := range taskDefinition.ContainerDefinitions {
		if *container.Name == "nginx" {
			nginxFound = true
			assert.Equal(t, 0, len(container.Secrets))
		}
	}
	assert.True(t, nginxFound)

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

	terraformRegistrySecretOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./environment_secrets",
		Vars: map[string]interface{}{
			"organization_id":     os.Getenv("ORGANIZATION_ID"),
			"environment_id":      os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":        os.Getenv("APTIBLE_HOST"),
			"plain_secret_1_name": "secret1",
			"plain_secret_1":      "dont expose",
			"plain_secret_2_name": "secret2",
			"plain_secret_2":      "asuperdupersecret",
			"json_secrets":        map[string]string{"secret1": "value1", "secret2": "value2", "secret3": "not used"},
		},
	})
	terraform.InitAndApply(t, terraformRegistrySecretOptions)
	defer terraform.Destroy(t, terraformRegistrySecretOptions)
	plainSecret1Arn := terraform.Output(t, terraformRegistrySecretOptions, "plain_secret_1_arn")
	plainSecret2Arn := terraform.Output(t, terraformRegistrySecretOptions, "plain_secret_2_arn")
	jsonSecretArn := terraform.Output(t, terraformRegistrySecretOptions, "json_secret_arn")

	terraformUpdateOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars: map[string]interface{}{
			"organization_id":   os.Getenv("ORGANIZATION_ID"),
			"environment_id":    os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":      os.Getenv("APTIBLE_HOST"),
			"dns_account_id":    os.Getenv("DNS_AWS_ACCOUNT_ID"),
			"ecs_name":          "ecs-pub-web-test-secrets",
			"container_command": []string{"nginx", "-g", "daemon off;"},
			"container_image":   "nginx",
			"container_port":    80,
			"container_name":    "nginx",
			"is_public":         true,
			"is_ecr_image":      false,
			"vpc_name":          "testecs-secrets-vpc",
			"domain":            "aptible-cloud-staging.com",
			"subdomain":         "test-ecs-integration",
			"environment_secrets": map[string]interface{}{
				"SECRET1": map[string]string{
					"secret_arn":      plainSecret1Arn,
					"secret_json_key": "",
				},
				"SECRET2": map[string]string{
					"secret_arn":      plainSecret2Arn,
					"secret_json_key": "",
				},
				"SECRET3": map[string]string{
					"secret_arn":      jsonSecretArn,
					"secret_json_key": "secret1",
				},
				"SECRET4": map[string]string{
					"secret_arn":      jsonSecretArn,
					"secret_json_key": "secret2",
				},
			},
		},
	})
	terraform.InitAndApply(t, terraformUpdateOptions)

	vpUpdateId := terraform.Output(t, terraformUpdateOptions, "vpc_id")
	assert.Equal(t, vpUpdateId, vpcId)
	vpcUpdateAsset, vpcUpdateAws, vpcUpdateErr := getAptibleAndAWSVPCs(t, ctx, c, vpUpdateId, "testecs-secrets-vpc")
	if assert.NoError(t, vpcUpdateErr) {
		assertCommonVpc(t, vpUpdateId, vpcUpdateAsset, vpcUpdateAws)
	}

	ecsUpdateWebId := terraform.Output(t, terraformUpdateOptions, "ecs_web_id")
	assert.Equal(t, ecsUpdateWebId, ecsWebId)
	ecsUpdateWebAsset, ecsUpdateClusterAws, ecsUpdateServiceAws, ecsUpdateErr := getAptibleAndAWSECSServiceAndCluster(t, ctx, c, ecsUpdateWebId, "ecs-pub-web-test-secrets-web-cluster", "ecs-pub-web-test-secrets")
	if assert.NoError(t, ecsUpdateErr) {
		assertCommonEcs(t, ecsUpdateWebId, ecsUpdateWebAsset, ecsUpdateClusterAws, ecsUpdateServiceAws)
	}
	assert.Equal(t, int64(1), *ecsUpdateServiceAws.DesiredCount)
	taskDefinitionUpdate := terratest_aws.GetEcsTaskDefinition(t, "us-east-1", *ecsUpdateServiceAws.TaskDefinition)
	assert.Equal(t, "awsvpc", *taskDefinitionUpdate.NetworkMode)
	assert.Equal(t, 2, len(taskDefinitionUpdate.ContainerDefinitions)) // ssl proxy and service container
	var nginxUpdateFound bool
	for _, container := range taskDefinitionUpdate.ContainerDefinitions {
		nginxUpdateFound = true
		if *container.Name == "nginx" {
			assert.Equal(t, 4, len(container.Secrets))
		}
	}
	assert.True(t, nginxUpdateFound)

	ecsUpdateLoadBalancerUrl := terraform.Output(t, terraformUpdateOptions, "loadbalancer_url")
	ecsUpdateLbGet, ecsUpdateLbGetErr := insecureHttpClient().Get(fmt.Sprintf("https://%s", ecsUpdateLoadBalancerUrl))
	if assert.NoError(t, ecsUpdateLbGetErr) {
		assert.EqualValues(t, ecsUpdateLbGet.StatusCode, 200)
	}

	ecsUpdateServiceUrl := terraform.Output(t, terraformUpdateOptions, "web_url")
	ecsUpdateUrlGet, ecsUpdateUrlGetErr := http.Get(fmt.Sprintf("https://%s", ecsUpdateServiceUrl))
	if assert.NoError(t, ecsUpdateUrlGetErr) {
		assert.EqualValues(t, ecsUpdateUrlGet.StatusCode, 200)
	}

	terraformRevertOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars:         initialVariables,
	})
	terraform.InitAndApply(t, terraformRevertOptions)

	vpcRevertId := terraform.Output(t, terraformRevertOptions, "vpc_id")
	assert.Equal(t, vpcRevertId, vpcId)
	vpcRevertAsset, vpcRevertAws, vpcRevertErr := getAptibleAndAWSVPCs(t, ctx, c, vpcRevertId, "testecs-secrets-vpc")
	if assert.NoError(t, vpcRevertErr) {
		assertCommonVpc(t, vpcRevertId, vpcRevertAsset, vpcRevertAws)
	}

	ecsRevertWebId := terraform.Output(t, terraformRevertOptions, "ecs_web_id")
	assert.Equal(t, ecsRevertWebId, ecsWebId)
	ecsRevertWebAsset, ecsRevertClusterAws, ecsRevertServiceAws, ecsRevertErr := getAptibleAndAWSECSServiceAndCluster(t, ctx, c, ecsRevertWebId, "ecs-pub-web-test-secrets-web-cluster", "ecs-pub-web-test-secrets")
	if assert.NoError(t, ecsRevertErr) {
		assertCommonEcs(t, ecsRevertWebId, ecsRevertWebAsset, ecsRevertClusterAws, ecsRevertServiceAws)
	}
	assert.Equal(t, int64(1), *ecsRevertServiceAws.DesiredCount)
	taskDefinitionRevert := terratest_aws.GetEcsTaskDefinition(t, "us-east-1", *ecsRevertServiceAws.TaskDefinition)
	assert.Equal(t, "awsvpc", *taskDefinitionRevert.NetworkMode)
	assert.Equal(t, 2, len(taskDefinitionRevert.ContainerDefinitions)) // ssl proxy and service container
	var nginxRevertFound bool
	for _, container := range taskDefinitionRevert.ContainerDefinitions {
		nginxRevertFound = true
		if *container.Name == "nginx" {
			assert.Equal(t, 0, len(container.Secrets))
		}
	}
	assert.True(t, nginxRevertFound)

	ecsRevertLoadBalancerUrl := terraform.Output(t, terraformRevertOptions, "loadbalancer_url")
	ecsRevertLbGet, ecsRevertLbGetErr := insecureHttpClient().Get(fmt.Sprintf("https://%s", ecsRevertLoadBalancerUrl))
	if assert.NoError(t, ecsRevertLbGetErr) {
		assert.EqualValues(t, ecsRevertLbGet.StatusCode, 200)
	}

	ecsRevertServiceUrl := terraform.Output(t, terraformRevertOptions, "web_url")
	ecsRevertUrlGet, ecsRevertUrlGetErr := http.Get(fmt.Sprintf("https://%s", ecsRevertServiceUrl))
	if assert.NoError(t, ecsRevertUrlGetErr) {
		assert.EqualValues(t, ecsRevertUrlGet.StatusCode, 200)
	}
}

func TestECSWebChangeContainer(t *testing.T) {
	initialVariables := map[string]interface{}{
		"organization_id":   os.Getenv("ORGANIZATION_ID"),
		"environment_id":    os.Getenv("ENVIRONMENT_ID"),
		"aptible_host":      os.Getenv("APTIBLE_HOST"),
		"dns_account_id":    os.Getenv("DNS_AWS_ACCOUNT_ID"),
		"ecs_name":          "ecs-pub-web-test-container",
		"container_command": []string{"nginx", "-g", "daemon off;"},
		"container_image":   "nginx",
		"container_port":    80,
		"container_name":    "nginx",
		"is_public":         true,
		"is_ecr_image":      false,
		"vpc_name":          "testecs-container-vpc",
		"domain":            "aptible-cloud-staging.com",
		"subdomain":         "test-ecs-integration",
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars:         initialVariables,
	})

	defer cleanupAndAssert(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	c := client.NewClient(
		true,
		os.Getenv("APTIBLE_HOST"),
		os.Getenv("APTIBLE_TOKEN"),
	)
	ctx := context.Background()

	aptibleAccountId := terraform.Output(t, terraformOptions, "aptible_aws_account_id")
	aptibleAccountRole := fmt.Sprintf("arn:aws:iam::%s:role/OrganizationAccountAccessRole", aptibleAccountId)
	os.Setenv(terratest_aws.AuthAssumeRoleEnvVar, aptibleAccountRole)

	vpcId := terraform.Output(t, terraformOptions, "vpc_id")
	vpcAsset, vpcAws, vpcErr := getAptibleAndAWSVPCs(t, ctx, c, vpcId, "testecs-container-vpc")
	if assert.NoError(t, vpcErr) {
		assertCommonVpc(t, vpcId, vpcAsset, vpcAws)
	}

	ecsWebId := terraform.Output(t, terraformOptions, "ecs_web_id")
	ecsWebAsset, ecsClusterAws, ecsServiceAws, ecsErr := getAptibleAndAWSECSServiceAndCluster(t, ctx, c, ecsWebId, "ecs-pub-web-test-container-web-cluster", "ecs-pub-web-test-container")
	if assert.NoError(t, ecsErr) {
		assertCommonEcs(t, ecsWebId, ecsWebAsset, ecsClusterAws, ecsServiceAws)
	}
	assert.Equal(t, int64(1), *ecsServiceAws.DesiredCount)
	taskDefinition := terratest_aws.GetEcsTaskDefinition(t, "us-east-1", *ecsServiceAws.TaskDefinition)
	assert.Equal(t, "awsvpc", *taskDefinition.NetworkMode)
	assert.Equal(t, 2, len(taskDefinition.ContainerDefinitions)) // ssl proxy and service container
	var nginxFound bool
	for _, container := range taskDefinition.ContainerDefinitions {
		if *container.Name == "nginx" {
			nginxFound = true
			assert.Equal(t, "nginx", *container.Image)
			assert.Equal(t, []*string{
				aws.String("nginx"),
				aws.String("-g"),
				aws.String("daemon off;"),
			}, container.Command)
		}
	}
	assert.True(t, nginxFound)

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

	terraformUpdateOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars: map[string]interface{}{
			"organization_id":   os.Getenv("ORGANIZATION_ID"),
			"environment_id":    os.Getenv("ENVIRONMENT_ID"),
			"aptible_host":      os.Getenv("APTIBLE_HOST"),
			"dns_account_id":    os.Getenv("DNS_AWS_ACCOUNT_ID"),
			"ecs_name":          "ecs-pub-web-test-container",
			"container_command": []string{"nginx-debug", "-g", "daemon off;"},
			"container_image":   "nginx:alpine",
			"container_port":    80,
			"container_name":    "nginx",
			"is_public":         true,
			"is_ecr_image":      false,
			"vpc_name":          "testecs-container-vpc",
			"domain":            "aptible-cloud-staging.com",
			"subdomain":         "test-ecs-integration",
		},
	})
	terraform.InitAndApply(t, terraformUpdateOptions)

	vpUpdateId := terraform.Output(t, terraformUpdateOptions, "vpc_id")
	assert.Equal(t, vpUpdateId, vpcId)
	vpcUpdateAsset, vpcUpdateAws, vpcUpdateErr := getAptibleAndAWSVPCs(t, ctx, c, vpUpdateId, "testecs-container-vpc")
	if assert.NoError(t, vpcUpdateErr) {
		assertCommonVpc(t, vpUpdateId, vpcUpdateAsset, vpcUpdateAws)
	}

	ecsUpdateWebId := terraform.Output(t, terraformUpdateOptions, "ecs_web_id")
	assert.Equal(t, ecsUpdateWebId, ecsWebId)
	ecsUpdateWebAsset, ecsUpdateClusterAws, ecsUpdateServiceAws, ecsUpdateErr := getAptibleAndAWSECSServiceAndCluster(t, ctx, c, ecsUpdateWebId, "ecs-pub-web-test-container-web-cluster", "ecs-pub-web-test-container")
	if assert.NoError(t, ecsUpdateErr) {
		assertCommonEcs(t, ecsUpdateWebId, ecsUpdateWebAsset, ecsUpdateClusterAws, ecsUpdateServiceAws)
	}
	assert.Equal(t, int64(1), *ecsUpdateServiceAws.DesiredCount)
	taskDefinitionUpdate := terratest_aws.GetEcsTaskDefinition(t, "us-east-1", *ecsUpdateServiceAws.TaskDefinition)
	assert.Equal(t, "awsvpc", *taskDefinitionUpdate.NetworkMode)
	assert.Equal(t, 2, len(taskDefinitionUpdate.ContainerDefinitions)) // ssl proxy and service container
	var nginxUpdateFound bool
	for _, container := range taskDefinitionUpdate.ContainerDefinitions {
		nginxUpdateFound = true
		if *container.Name == "nginx" {
			assert.Equal(t, "nginx:alpine", *container.Image)
			assert.Equal(t, []*string{
				aws.String("nginx-debug"),
				aws.String("-g"),
				aws.String("daemon off;"),
			}, container.Command)
		}
	}
	assert.True(t, nginxUpdateFound)

	ecsUpdateLoadBalancerUrl := terraform.Output(t, terraformUpdateOptions, "loadbalancer_url")
	ecsUpdateLbGet, ecsUpdateLbGetErr := insecureHttpClient().Get(fmt.Sprintf("https://%s", ecsUpdateLoadBalancerUrl))
	if assert.NoError(t, ecsUpdateLbGetErr) {
		assert.EqualValues(t, ecsUpdateLbGet.StatusCode, 200)
	}

	ecsUpdateServiceUrl := terraform.Output(t, terraformUpdateOptions, "web_url")
	ecsUpdateUrlGet, ecsUpdateUrlGetErr := http.Get(fmt.Sprintf("https://%s", ecsUpdateServiceUrl))
	if assert.NoError(t, ecsUpdateUrlGetErr) {
		assert.EqualValues(t, ecsUpdateUrlGet.StatusCode, 200)
	}

	terraformRevertOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
		Vars:         initialVariables,
	})
	terraform.InitAndApply(t, terraformRevertOptions)

	vpcRevertId := terraform.Output(t, terraformRevertOptions, "vpc_id")
	assert.Equal(t, vpcRevertId, vpcId)
	vpcRevertAsset, vpcRevertAws, vpcRevertErr := getAptibleAndAWSVPCs(t, ctx, c, vpcRevertId, "testecs-container-vpc")
	if assert.NoError(t, vpcRevertErr) {
		assertCommonVpc(t, vpcRevertId, vpcRevertAsset, vpcRevertAws)
	}

	ecsRevertWebId := terraform.Output(t, terraformRevertOptions, "ecs_web_id")
	assert.Equal(t, ecsRevertWebId, ecsWebId)
	ecsRevertWebAsset, ecsRevertClusterAws, ecsRevertServiceAws, ecsRevertErr := getAptibleAndAWSECSServiceAndCluster(t, ctx, c, ecsRevertWebId, "ecs-pub-web-test-container-web-cluster", "ecs-pub-web-test-container")
	if assert.NoError(t, ecsRevertErr) {
		assertCommonEcs(t, ecsRevertWebId, ecsRevertWebAsset, ecsRevertClusterAws, ecsRevertServiceAws)
	}
	assert.Equal(t, int64(1), *ecsRevertServiceAws.DesiredCount)
	taskDefinitionRevert := terratest_aws.GetEcsTaskDefinition(t, "us-east-1", *ecsRevertServiceAws.TaskDefinition)
	assert.Equal(t, "awsvpc", *taskDefinitionRevert.NetworkMode)
	assert.Equal(t, 2, len(taskDefinitionRevert.ContainerDefinitions)) // ssl proxy and service container
	var nginxRevertFound bool
	for _, container := range taskDefinitionRevert.ContainerDefinitions {
		nginxRevertFound = true
		if *container.Name == "nginx" {
			assert.Equal(t, "nginx", *container.Image)
			assert.Equal(t, []*string{
				aws.String("nginx"),
				aws.String("-g"),
				aws.String("daemon off;"),
			}, container.Command)
		}
	}
	assert.True(t, nginxRevertFound)

	ecsRevertLoadBalancerUrl := terraform.Output(t, terraformRevertOptions, "loadbalancer_url")
	ecsRevertLbGet, ecsRevertLbGetErr := insecureHttpClient().Get(fmt.Sprintf("https://%s", ecsRevertLoadBalancerUrl))
	if assert.NoError(t, ecsRevertLbGetErr) {
		assert.EqualValues(t, ecsRevertLbGet.StatusCode, 200)
	}

	ecsRevertServiceUrl := terraform.Output(t, terraformRevertOptions, "web_url")
	ecsRevertUrlGet, ecsRevertUrlGetErr := http.Get(fmt.Sprintf("https://%s", ecsRevertServiceUrl))
	if assert.NoError(t, ecsRevertUrlGetErr) {
		assert.EqualValues(t, ecsRevertUrlGet.StatusCode, 200)
	}
}
