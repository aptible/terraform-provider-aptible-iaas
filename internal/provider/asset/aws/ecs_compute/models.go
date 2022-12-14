package ecscompute

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	assetutil "github.com/aptible/terraform-provider-aptible-iaas/internal/provider/asset/util"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/util"
)

var resourceTypeName = "_aws_ecs_compute"
var resourceDescription = "ECS compute resource"

type Env struct {
	SecretArn     types.String `tfsdk:"secret_arn" json:"secret_arn"`
	SecretJsonKey types.String `tfsdk:"secret_json_key" json:"secret_json_key"`
}

type EnvJson struct {
	EnvVar        string `json:"environment_variable"`
	SecretArn     string `json:"secret_arn"`
	SecretJsonKey string `json:"secret_json_key"`
}

// TODO - autogenerated
type ResourceModel struct {
	Id             types.String `tfsdk:"id" json:"id"`
	AssetVersion   types.String `tfsdk:"asset_version" json:"asset_version"`
	EnvironmentId  types.String `tfsdk:"environment_id" json:"environment_id"`
	OrganizationId types.String `tfsdk:"organization_id" json:"organization_id"`
	Status         types.String `tfsdk:"status" json:"status"`

	VpcName                    types.String   `tfsdk:"vpc_name" json:"vpc_name"`
	Name                       types.String   `tfsdk:"name" json:"name"`
	EnvironmentSecrets         map[string]Env `tfsdk:"environment_secrets" json:"environment_secrets"`
	ContainerName              types.String   `tfsdk:"container_name" json:"container_name"`
	ContainerPort              types.Number   `tfsdk:"container_port" json:"container_port"`
	ContainerImage             types.String   `tfsdk:"container_image" json:"container_image"`
	ContainerCommand           []types.String `tfsdk:"container_command" json:"container_command"`
	ConnectsTo                 types.List     `tfsdk:"connects_to"`
	ContainerRegistrySecretArn types.String   `tfsdk:"container_registry_secret_arn"`
	IsEcrImage                 types.Bool     `tfsdk:"is_ecr_image"`
	WaitForSteadyState         types.Bool     `tfsdk:"wait_for_steady_state"`
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
		Computed: true,
	},
	"name": {
		Type:     types.StringType,
		Required: true,
	},
	"container_name": {
		Type:     types.StringType,
		Required: true,
	},
	"container_image": {
		Type:     types.StringType,
		Required: true,
	},
	"container_registry_secret_arn": {
		Type:     types.StringType,
		Optional: true,
	},
	"container_port": {
		Type:     types.NumberType,
		Required: true,
	},
	"container_command": {
		Type:     types.ListType{ElemType: types.StringType},
		Required: true,
	},
	"connects_to": {
		Type:     types.ListType{ElemType: types.StringType},
		Optional: true,
	},
	"environment_secrets": {
		Required: true,
		Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
			"secret_arn": {
				Type:     types.StringType,
				Required: true,
			},
			"secret_json_key": {
				Type:     types.StringType,
				Required: true,
			},
		}),
	},
	"is_ecr_image": {
		Type:     types.BoolType,
		Optional: true,
		Computed: true, // if unset, will default to false returned by backend
	},
	"wait_for_steady_state": {
		Type:     types.BoolType,
		Optional: true,
		Computed: true, // if unset, will default to false returned by backend
	},
}

func planToAssetInput(ctx context.Context, plan ResourceModel) (cac.AssetInput, error) {
	cmd := []string{}
	for _, c := range plan.ContainerCommand {
		cmd = append(cmd, c.Value)
	}
	secrets := []EnvJson{}
	for k, v := range plan.EnvironmentSecrets {
		secrets = append(secrets, EnvJson{
			EnvVar:        k,
			SecretArn:     v.SecretArn.Value,
			SecretJsonKey: v.SecretJsonKey.Value,
		})
	}

	params := map[string]interface{}{
		"vpc_name":            plan.VpcName.Value,
		"name":                plan.Name.Value,
		"container_name":      plan.ContainerName.Value,
		"container_image":     plan.ContainerImage.Value,
		"container_port":      plan.ContainerPort.Value,
		"container_command":   cmd,
		"environment_secrets": secrets,
	}

	if !plan.ContainerRegistrySecretArn.IsNull() && !plan.ContainerRegistrySecretArn.IsUnknown() {
		params["container_registry_secret_arn"] = plan.ContainerRegistrySecretArn.Value
	}

	if !plan.IsEcrImage.IsNull() && !plan.IsEcrImage.IsUnknown() {
		params["is_ecr_image"] = plan.IsEcrImage.Value
	}

	if !plan.WaitForSteadyState.IsNull() && !plan.WaitForSteadyState.IsUnknown() {
		params["wait_for_steady_state"] = plan.WaitForSteadyState.Value
	}

	// TODO HACK: https://aptible.slack.com/archives/C03C2STPTDX/p1664478414991299
	input := cac.AssetInput{
		Asset:           client.CompileAsset("aws", "ecs_compute_service", assetutil.DefaultAssetVersion),
		AssetVersion:    assetutil.DefaultAssetVersion,
		AssetParameters: params,
	}

	if !plan.ConnectsTo.IsNull() && !plan.ConnectsTo.IsUnknown() {
		connect := []string{}
		_ = plan.ConnectsTo.ElementsAs(ctx, &connect, false)
		input.ConnectsTo = connect
	}

	return input, nil
}

func assetOutputToPlan(ctx context.Context, plan ResourceModel, output *cac.AssetOutput) (*ResourceModel, error) {
	cmd := []types.String{}
	cmdList := output.CurrentAssetParameters.Data["container_command"].([]interface{})
	for _, c := range cmdList {
		cmd = append(cmd, types.String{Value: c.(string)})
	}

	/* connect := []attr.Value{}
	for _, c := range output.ConnectsTo {
		connect = append(connect, types.String{Value: c})
	}
	connectsTo := types.List{Elems: connect, ElemType: types.StringType}
	if len(connect) == 0 {
		connectsTo.Null = true
	} */
	// TODO: HACK we are not keeping what the API sends us because the API changes the
	// order which causes terraform to error
	connectsTo := plan.ConnectsTo

	// TODO: figure out how to not need an intermediate struct for marshal/unmarshal
	secretsJson := []EnvJson{}
	bts, err := json.Marshal(output.CurrentAssetParameters.Data["environment_secrets"])
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bts, &secretsJson)
	if err != nil {
		return nil, err
	}

	secrets := map[string]Env{}
	for _, v := range secretsJson {
		secrets[v.EnvVar] = Env{
			SecretArn:     types.String{Value: v.SecretArn},
			SecretJsonKey: types.String{Value: v.SecretJsonKey},
		}
	}

	port := output.CurrentAssetParameters.Data["container_port"].(float64)

	model := &ResourceModel{
		Id:                         types.String{Value: output.Id},
		AssetVersion:               types.String{Value: output.AssetVersion},
		EnvironmentId:              types.String{Value: output.Environment.Id},
		OrganizationId:             types.String{Value: output.Environment.Organization.Id},
		Status:                     types.String{Value: string(output.Status)},
		VpcName:                    types.String{Value: output.CurrentAssetParameters.Data["vpc_name"].(string)},
		Name:                       types.String{Value: output.CurrentAssetParameters.Data["name"].(string)},
		ContainerName:              types.String{Value: output.CurrentAssetParameters.Data["container_name"].(string)},
		ContainerPort:              types.Number{Value: big.NewFloat(port)},
		ContainerImage:             types.String{Value: output.CurrentAssetParameters.Data["container_image"].(string)},
		ContainerRegistrySecretArn: util.StringVal(output.CurrentAssetParameters.Data["container_registry_secret_arn"]),
		ContainerCommand:           cmd,
		ConnectsTo:                 connectsTo,
		EnvironmentSecrets:         secrets,
		IsEcrImage:                 util.BoolVal(output.CurrentAssetParameters.Data["is_ecr_image"]),
		WaitForSteadyState:         util.BoolVal(output.CurrentAssetParameters.Data["wait_for_steady_state"]),
	}

	return model, nil
}
