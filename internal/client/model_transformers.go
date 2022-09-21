package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"

	cloud_api_client "github.com/aptible/cloud-api-clients/clients/go"
)

const DELIMITER = "__"

var TOP_LEVEL_KEYS = []string{
	"id",
	"asset_platform",
	"asset_type",
	"asset_version",
	"environment_id",
	"organization_id",
	"status",
	// terraform internal keys
	"tf_provider_addr",
	"tf_resource_type",
	"@caller",
	"@module",
	"timestamp",
}

/**
name="map[Null:false Unknown:false Value:my_null]" tf_resource_type=aptible_null_simple tf_rpc=ApplyResourceChange asset_type="map[Null:false Unknown:false Value:simple]" @caller=/Users/madhu/work/terraform-provider-aptible-iaas/internal/client/model_transformers.go:40 @module=aptible_iaas asset_version="map[Null:false Unknown:false Value:latest]" id="map[Null:false Unknown:true Value:]" organization_id="map[Null:false Unknown:false Value:2253ae98-d65a-4180-aceb-8419b7416677]" status="map[Null:false Unknown:true Value:]" tf_provider_addr=aptible.com/aptible/aptible-iaas tf_req_id=e6b2222c-24d2-84fe-2a29-24384c2cead0 asset_platform="map[Null:false Unknown:false Value:null]" environment_id="map[Null:false Unknown:false Value:238930f4-0750-4f55-b43c-e1a11c437e23]" timestamp=2022-09-21T18:31:09.519-0400
2022-09-21T18:31:09.519-0400 [INFO]  provider.terraform-provider-aptible-iaas_0.0.0+local_darwin_arm64: Using these asset para
*/
// exclude top level keys

func PopulateClientAssetInputForCreate(ctx context.Context, input []byte, assetName, cloud, version string) (*cloud_api_client.AssetInput, error) {
	allOutput := make(map[string]interface{})
	assetParameters := make(map[string]interface{})
	if err := json.Unmarshal(input, &allOutput); err != nil {
		return nil, err
	}

	tflog.Trace(ctx, "Sending following data to backend", allOutput)
	for k, v := range allOutput {
		for _, excludedKey := range TOP_LEVEL_KEYS {
			if excludedKey == k {
				continue
			}
			switch value := v.(type) {
			// TODO - look at multiple
			case types.String:
				assetParameters[k] = value.Value
			case types.Bool:
				assetParameters[k] = value.Value
			default:
				tflog.Warn(ctx, fmt.Sprintf("unexpected type (%s) on key (%s)", value, k))
			}
		}
	}

	tflog.Info(ctx, "Using these asset parameters", assetParameters)

	return &cloud_api_client.AssetInput{
		Asset: fmt.Sprintf(
			"%s%s%s%s%s",
			cloud,
			DELIMITER,
			assetName,
			DELIMITER,
			version,
		),
		AssetVersion:    version,
		AssetParameters: assetParameters,
	}, nil
}

// take a set of inputs, describe the asset, and safely interpolate for recreate
func PopulateClientAssetInputForUpdate(ctx context.Context, cloudApiAssetState *cloud_api_client.AssetOutput, input []byte, assetName, cloud, version string) (*cloud_api_client.AssetInput, error) {
	creationInput, err := PopulateClientAssetInputForCreate(ctx, input, assetName, cloud, version)
	if err != nil {
		return nil, err
	}

	maps.Copy(creationInput.AssetParameters, cloudApiAssetState.CurrentAssetParameters.Data)

	return creationInput, nil
}
