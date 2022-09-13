package aptible

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cloud_api_client "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/aptible/models"
)

const DELIMITER = "__"

type resourceAssetType struct{}

var assetSchema = map[string]tfsdk.Attribute{
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
	"parameters": {
		Type:     types.StringType,
		Optional: true,
	},
}

func (r resourceAssetType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: assetSchema,
	}, nil
}

func (r resourceAssetType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceAsset{
		p: *(p.(*provider)),
	}, nil
}

type resourceAsset struct {
	p provider
}

func (r resourceAsset) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var asset models.Asset
	diags := req.Plan.Get(ctx, &asset)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var assetParametersJson map[string]interface{}
	err := json.Unmarshal([]byte(asset.Parameters.String()), &assetParametersJson)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to generate json from asset parameters",
			fmt.Sprintf("JSON marshalling error - %s", err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Creating asset")
	createdAsset, err := r.p.client.CreateAsset(
		asset.OrganizationId.String(),
		asset.EnvironmentId.String(),
		cloud_api_client.AssetInput{
			Asset: fmt.Sprintf(
				"%s%s%s%s%s%s",
				asset.AssetPlatform,
				DELIMITER,
				asset.AssetType,
				DELIMITER,
				asset.AssetVersion,
				DELIMITER,
			),
			AssetParameters: assetParametersJson,
			AssetVersion:    asset.AssetVersion.String(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating asset",
			"Could not create asset, unexpected error: "+err.Error(),
		)
		return
	}
	// for more information on logging from providers, refer to
	// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog
	tflog.Trace(ctx, "created asset", map[string]interface{}{"id": createdAsset.Id, "status": createdAsset.Status})

	stringAssetParameters, err := json.Marshal(createdAsset.CurrentAssetParameters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshaling asset parameters",
			"Could not marshal asset parameters json, unexpected error: "+err.Error(),
		)
		return
	}

	result := models.Asset{
		AssetBase: models.AssetBase{
			Id:             types.String{Value: createdAsset.Id},
			EnvironmentId:  types.String{Value: createdAsset.Environment.Id},
			OrganizationId: types.String{Value: createdAsset.Environment.Organization.Id},
		},
		Parameters: types.String{Value: string(stringAssetParameters)},
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// todo - waiter
}

// Read resource information
func (r resourceAsset) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state models.Asset
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	assetClientOutput, err := r.p.client.DescribeAsset(state.OrganizationId.String(), state.EnvironmentId.String(), state.Id.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading asset",
			"Could not read ID "+state.Id.String()+": "+err.Error(),
		)
		return
	}

	stringAssetParameters, err := json.Marshal(assetClientOutput.CurrentAssetParameters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshaling asset parameters",
			"Could not marshal asset parameters json, unexpected error: "+err.Error(),
		)
		return
	}

	// interpolate retrieved asset info with existing state
	asset := models.Asset{
		AssetBase: models.AssetBase{
			Id:             types.String{Value: assetClientOutput.Id},
			EnvironmentId:  types.String{Value: assetClientOutput.Environment.Id},
			OrganizationId: types.String{Value: assetClientOutput.Environment.Organization.Id},
		},
		Parameters: types.String{Value: string(stringAssetParameters)},
	}

	// Set state
	diags = resp.State.Set(ctx, &asset)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceAsset) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan values
	var plan models.Asset
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state and compare against remote
	assetInCloudApi, err := r.p.client.DescribeAsset(plan.OrganizationId.String(), plan.EnvironmentId.String(), plan.Id.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error update asset",
			"Could not update asset id "+assetInCloudApi.Id+": "+err.Error(),
		)
		return
	}

	// request update
	result, err := r.p.client.UpdateAsset(
		assetInCloudApi.Id,
		assetInCloudApi.Environment.Id,
		assetInCloudApi.Environment.Organization.Id,
		cloud_api_client.AssetInput{
			Asset: fmt.Sprintf(
				"%s%s%s%s%s%s",
				plan.AssetPlatform,
				DELIMITER,
				plan.AssetType,
				DELIMITER,
				plan.AssetVersion,
				DELIMITER,
			),
			AssetParameters: assetInCloudApi.CurrentAssetParameters.Data,
			AssetVersion:    plan.AssetVersion.String(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error requesting update from cloud api",
			"Could not marshal asset parameters json, unexpected error: "+err.Error(),
		)
		return
	}

	// generate json from update
	assetParametersFromUpdate, err := json.Marshal(result.CurrentAssetParameters.Data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshaling asset parameters",
			"Could not marshal asset parameters json, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state
	diags = resp.State.Set(ctx, models.Asset{
		AssetBase: models.AssetBase{
			Id:             types.String{Value: result.Id},
			EnvironmentId:  types.String{Value: result.Environment.Id},
			OrganizationId: types.String{Value: result.Environment.Organization.Id},
		},
		Parameters: types.String{Value: string(assetParametersFromUpdate)},
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// todo - waiter
}

// Delete resource
func (r resourceAsset) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state models.Asset
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete asset by calling API
	err := r.p.client.DestroyAsset(state.OrganizationId.String(), state.EnvironmentId.String(), state.Id.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting asset",
			"Could not delete asset id "+state.Id.String()+": "+err.Error(),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceAsset) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
