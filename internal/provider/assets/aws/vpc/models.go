package vpc

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cloud_api_client "github.com/aptible/cloud-api-clients/clients/go"
)

// TODO - autogenerated
type VPC struct {
	Id             types.String `tfsdk:"id" json:"id"`
	AssetVersion   types.String `tfsdk:"asset_version" json:"asset_version"`
	EnvironmentId  types.String `tfsdk:"environment_id" json:"environment_id"`
	OrganizationId types.String `tfsdk:"organization_id" json:"organization_id"`
	Status         types.String `tfsdk:"status" json:"status"`

	Name types.String `tfsdk:"name" json:"name"`
}

var AssetSchema = map[string]tfsdk.Attribute{
	"id": {
		Description: "A valid asset id",
		Type:        types.StringType,
		Computed:    true,
	},
	"status": {
		Type:     types.StringType,
		Computed: true,
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
	"asset_version": {
		Type:     types.StringType,
		Required: true,
	},
	"name": {
		Type:     types.StringType,
		Required: true,
	},
}

func GenerateResourceFromAssetOutput(output *cloud_api_client.AssetOutput) (*VPC, error) {
	vpc := &VPC{
		Id:             types.String{Value: output.Id},
		AssetVersion:   types.String{Value: output.AssetVersion},
		EnvironmentId:  types.String{Value: output.Environment.Id},
		OrganizationId: types.String{Value: output.Environment.Organization.Id},
		Status:         types.String{Value: string(output.Status)},
		Name:           types.String{Value: output.CurrentAssetParameters.Data["name"].(string)},
	}

	return vpc, nil
}
