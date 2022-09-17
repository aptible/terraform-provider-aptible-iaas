package client

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"

	cloud_api_client "github.com/aptible/cloud-api-clients/clients/go"
)

const DELIMITER = "__"

type MirroredAssetBase struct {
	AssetVersion types.String
}

type AssetStructMirror struct {
	MirroredAssetBase // TODO - doing this to avoid circular deps for now
	AssetParameters   any
}

func PopulateClientAssetInputForCreate(input any, assetName, cloud string) (*cloud_api_client.AssetInput, error) {
	// all inputs have two categories: asset base, which are interpolated
	expectedInput := input.(AssetStructMirror)

	// convert to map[string] interface{}for asset parameter usage as it needs
	var assetParameters map[string]interface{}
	rawData, err := json.Marshal(expectedInput.AssetParameters)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rawData, &assetParameters); err != nil {
		return nil, err
	}

	return &cloud_api_client.AssetInput{
		Asset: fmt.Sprintf(
			"%s%s%s%s%s%s",
			cloud,
			DELIMITER,
			assetName,
			DELIMITER,
			expectedInput.AssetVersion,
		),
		AssetVersion:    expectedInput.AssetVersion.String(),
		AssetParameters: assetParameters,
	}, nil
}

// take a set of inputs, describe the asset, and safely interpolate for recreate
func PopulateClientAssetInputForUpdate(cloudApiAssetState *cloud_api_client.AssetOutput, input any, assetName, cloud string) (*cloud_api_client.AssetInput, error) {
	creationInput, err := PopulateClientAssetInputForCreate(input, assetName, cloud)
	if err != nil {
		return nil, err
	}

	maps.Copy(creationInput.AssetParameters, cloudApiAssetState.CurrentAssetParameters.Data)

	return creationInput, nil
}
