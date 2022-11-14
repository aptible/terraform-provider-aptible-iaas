package utils

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	tagging_types "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const maxPages = 10

// GetTaggedResources -
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
			resource_arn := *resource.ResourceARN
			isActive, err := verifyResourcesIsActive(ctx, resource_arn)

			if err != nil {
				return nil, err
			}

			if isActive {
				arnSlice = append(arnSlice, resource_arn)
			}
		}

		if response.PaginationToken == nil {
			// No more pages so we break.
			break
		}

		// Assign the PaginationToken pointer to page_token for the next loop
		params.PaginationToken = response.PaginationToken
	}

	return arnSlice, nil
}

func verifyResourcesIsActive(ctx context.Context, resource_arn string) (bool, error) {

	switch {
	case strings.Contains(resource_arn, "arn:aws:kms"):
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return false, err
		}

		c := kms.NewFromConfig(cfg)
		key, err := c.DescribeKey(ctx, &kms.DescribeKeyInput{KeyId: &resource_arn})

		if err != nil {
			return false, err
		}
		// If the deletion date does not point at nil the resource is still active.
		return key.KeyMetadata.DeletionDate != nil, nil

	case strings.Contains(resource_arn, "arn:aws:secretsmanager"):
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return true, err
		}

		c := secretsmanager.NewFromConfig(cfg)
		secret, err := c.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{SecretId: &resource_arn})

		if err != nil {
			return true, err
		}
		// If the deletion date does not point at nil the resource is still active.
		return secret.DeletedDate != nil, nil

	}

	// If we don't have a special case for the resource type assume it is active.
	return true, nil
}
