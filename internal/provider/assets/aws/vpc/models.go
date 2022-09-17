package vpc

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cloud_api_client "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/common"
)

type VPCAssetParameters struct {
	Name        types.String `tfsdk:"name" json:"name"`
	Description types.String `tfsdk:"description" json:"description"`
}

type VPC struct {
	common.AssetBase
	AssetParameters VPCAssetParameters
}

// TODO - freeze below maps with immutable or freeze libraries
var AssetSchema = map[string]tfsdk.Attribute{
	"id": {
		Description: "A valid asset id",
		Type:        types.StringType,
		Computed:    true,
	},
	"environment_id": {
		Description: "A valid environment id",
		Type:        types.StringType,
		Required:    true,
	},
	"organization_id": {
		Description: "A valid organization id",
		Type:        types.StringType,
		Required:    true,
	},
	"asset_platform": {
		Type:     types.StringType,
		Required: true,
	},
	"asset_type": {
		Type:     types.StringType,
		Required: true,
	},
	"asset_version": {
		Type:     types.StringType,
		Required: true,
	},
	"name": {
		Type:     types.StringType,
		Required: true,
	},
}

// TODO - turn this into a generic for all asset types?
func GenerateResourceFromAssetOutput(output *cloud_api_client.AssetOutput) (*VPC, error) {
	var assetParameters VPCAssetParameters
	rawData, err := json.Marshal(output.CurrentAssetParameters.Data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rawData, &assetParameters); err != nil {
		return nil, err
	}

	return &VPC{
		AssetBase: common.AssetBase{
			Id:             types.String{Value: output.Id},
			AssetVersion:   types.String{Value: output.AssetVersion},
			EnvironmentId:  types.String{Value: output.Environment.Id},
			OrganizationId: types.String{Value: output.Environment.Organization.Id},
			Status:         types.String{Value: string(output.Status)},
		},
		AssetParameters: assetParameters,
	}, nil
}
