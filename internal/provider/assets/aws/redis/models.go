package redis

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
)

var resourceTypeName = "_aws_redis"
var resourceDescription = "Redis resource"

// TODO - autogenerated
type ResourceModel struct {
	Id             types.String `tfsdk:"id" json:"id"`
	AssetVersion   types.String `tfsdk:"asset_version" json:"asset_version"`
	EnvironmentId  types.String `tfsdk:"environment_id" json:"environment_id"`
	OrganizationId types.String `tfsdk:"organization_id" json:"organization_id"`
	Status         types.String `tfsdk:"status" json:"status"`

	VpcName           types.String `tfsdk:"vpc_name" json:"vpc_name"`
	Name              types.String `tfsdk:"name" json:"name"`
	Description       types.String `tfsdk:"description" json:"description"`
	SnapshotWindow    types.String `tfsdk:"snapshot_window" json:"snapshot_window"`
	MaintenanceWindow types.String `tfsdk:"maintenance_window" json:"maintainence_window"`
	UriSecretArn      types.String `tfsdk:"uri_secret_arn" json:"elasticache_token_secret_arn"`
	SecretsKmsKeyArn  types.String `tfsdk:"secrets_kms_key_arn" json:"elasticache_token_kms_key_arn"`
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
	"vpc_name": {
		Description: "A valid vpc name",
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
	"description": {
		Type:     types.StringType,
		Required: true,
	},
	"snapshot_window": {
		Type:     types.StringType,
		Required: true,
	},
	"maintenance_window": {
		Type:     types.StringType,
		Required: true,
	},
	"uri_secret_arn": {
		Computed: true,
		Type:     types.StringType,
	},
	"secrets_kms_key_arn": {
		Computed: true,
		Type:     types.StringType,
	},
}

func planToAssetInput(plan ResourceModel) (cac.AssetInput, error) {
	input := cac.AssetInput{
		Asset:        client.CompileAsset("aws", "elasticache_redis", plan.AssetVersion.Value),
		AssetVersion: plan.AssetVersion.Value,
		AssetParameters: map[string]interface{}{
			"vpc_name":           plan.VpcName.Value,
			"name":               plan.Name.Value,
			"description":        plan.Description.Value,
			"snapshot_window":    plan.SnapshotWindow.Value,
			"maintenance_window": plan.MaintenanceWindow.Value,
		},
	}

	return input, nil
}

func assetOutputToPlan(output *cac.AssetOutput) (*ResourceModel, error) {
	outputs := *output.Outputs

	model := &ResourceModel{
		Id:                types.String{Value: output.Id},
		AssetVersion:      types.String{Value: output.AssetVersion},
		EnvironmentId:     types.String{Value: output.Environment.Id},
		OrganizationId:    types.String{Value: output.Environment.Organization.Id},
		Status:            types.String{Value: string(output.Status)},
		VpcName:           types.String{Value: output.CurrentAssetParameters.Data["vpc_name"].(string)},
		Name:              types.String{Value: output.CurrentAssetParameters.Data["name"].(string)},
		Description:       types.String{Value: output.CurrentAssetParameters.Data["description"].(string)},
		SnapshotWindow:    types.String{Value: output.CurrentAssetParameters.Data["snapshot_window"].(string)},
		MaintenanceWindow: types.String{Value: output.CurrentAssetParameters.Data["maintenance_window"].(string)},
		UriSecretArn:      types.String{Value: outputs["elasticache_token_secret_arn"].Data.(string)},
		SecretsKmsKeyArn:  types.String{Value: outputs["elasticache_token_kms_key_arn"].Data.(string)},
	}

	return model, nil
}
