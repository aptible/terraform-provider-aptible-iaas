package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
)

var resourceTypeName = "_aws_vpc"
var resourceDescription = "VPC resource"

// TODO - autogenerated
type ResourceModel struct {
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

func planToAssetInput(ctx context.Context, plan ResourceModel) (cac.AssetInput, error) {
	input := cac.AssetInput{
		Asset:        client.CompileAsset("aws", "vpc", plan.AssetVersion.Value),
		AssetVersion: plan.AssetVersion.Value,
		AssetParameters: map[string]interface{}{
			"name": plan.Name.Value,
		},
	}

	return input, nil
}

func assetOutputToPlan(ctx context.Context, output *cac.AssetOutput) (*ResourceModel, error) {
	vpc := &ResourceModel{
		Id:             types.String{Value: output.Id},
		AssetVersion:   types.String{Value: output.AssetVersion},
		EnvironmentId:  types.String{Value: output.Environment.Id},
		OrganizationId: types.String{Value: output.Environment.Organization.Id},
		Status:         types.String{Value: string(output.Status)},
		Name:           types.String{Value: output.CurrentAssetParameters.Data["name"].(string)},
	}

	return vpc, nil
}
