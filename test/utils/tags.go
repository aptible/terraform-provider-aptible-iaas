//package main
package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	tagging_types "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const maxPages = 10

// GetTaggedResources returns active resources with the environment and asset tags
func GetTaggedResources(ctx context.Context, environmentId, assetId string) ([]string, error) {

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	c := resourcegroupstaggingapi.NewFromConfig(cfg)

	var arnSlice []string

	params := &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []tagging_types.TagFilter{
			{Key: aws.String("aptible_asset_id"), Values: []string{assetId}},
			{Key: aws.String("aptible_environment_id"), Values: []string{environmentId}},
		},
	}

	for i := 0; i < maxPages; i++ {
		response, err := c.GetResources(ctx, params)
		if err != nil {
			return arnSlice, err
		}

		for _, resource := range response.ResourceTagMappingList {
			resourceARN := *resource.ResourceARN
			isActive, err := verifyResourcesIsActive(ctx, resourceARN)

			if err != nil {
				return nil, err
			}

			if isActive {
				arnSlice = append(arnSlice, resourceARN)
			}
		}

		if response.PaginationToken == nil {
			// No more pages so we break.
			break
		}

		if i+1 == maxPages {
			fmt.Println("Hit Max Pages when attempting to load all of the resources. Returning a truncated list.")
		}

		// Assign the PaginationToken pointer to page_token for the next loop
		params.PaginationToken = response.PaginationToken
	}

	return arnSlice, nil
}

// Verifies that a resource is actually active. This has edge cases for different resource types.
func verifyResourcesIsActive(ctx context.Context, resourceARN string) (bool, error) {

	instanceRegex := regexp.MustCompile(`arn:aws:ec2.*instance/(i-[a-f0-9]{17})`)

	switch {
	case strings.Contains(resourceARN, "arn:aws:kms"):
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return false, err
		}

		c := kms.NewFromConfig(cfg)
		key, err := c.DescribeKey(ctx, &kms.DescribeKeyInput{KeyId: &resourceARN})

		if err != nil {
			return false, err
		}
		// If the deletion date points at nil the resource is still active.
		return key.KeyMetadata.DeletionDate == nil, nil

	case strings.Contains(resourceARN, "arn:aws:secretsmanager"):
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return true, err
		}

		c := secretsmanager.NewFromConfig(cfg)
		secret, err := c.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{SecretId: &resourceARN})

		if err != nil {
			return true, err
		}
		// If the deletion date points at nil the resource is still active.
		return secret.DeletedDate == nil, nil

	case instanceRegex.MatchString(resourceARN):

		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return true, err
		}

		c := ec2.NewFromConfig(cfg)

		// Extract the instance ID since the APIs do not accept the ARN.
		// r := regexp.MustCompile(`instance/(i-[a-f0-9]{17})`)
		results := instanceRegex.FindStringSubmatch(resourceARN)

		if len(results) < 2 {
			fmt.Println(resourceARN)
			return true, fmt.Errorf("Unable to get instance id from %q", resourceARN)
		}

		instanceId := results[1]
		instanceIds := []string{instanceId}
		includeAll := true
		instance, err := c.DescribeInstanceStatus(ctx, &ec2.DescribeInstanceStatusInput{InstanceIds: instanceIds, IncludeAllInstances: &includeAll})

		if err != nil {
			return true, err
		}

		// The instances may disappear between the tagging resource call and the describe state call.
		// If it does disappear it means the instance is not active.
		if len(instance.InstanceStatuses) != 0 {
			return false, err
		}

		// If the instance code is not terminated than the instance is active.
		terminationCode := int32(48)
		return instance.InstanceStatuses[0].InstanceState.Code != &terminationCode, nil
	}

	// If we don't have a special case for the resource type assume it is active.
	return true, nil
}
