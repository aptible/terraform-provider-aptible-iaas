package simple

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/common"
)

type ResourceAssetType struct{}

func (r ResourceAssetType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: AssetSchema,
	}, nil
}

func (r ResourceAssetType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceAsset{
		p: *(p.(*common.Provider)),
	}, nil
}

type resourceAsset struct {
	p common.Provider
}

func (r resourceAsset) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var asset Simple
	diags := req.Plan.Get(ctx, &asset)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating asset", map[string]interface{}{"asset": asset})

	rawData, err := json.Marshal(asset)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling vpc plan",
			"Could not marshal vpc plan, unexpected error: "+err.Error(),
		)
		return
	}

	assetInput, err := client.PopulateClientAssetInputForCreate(ctx, rawData, "vpc", "aws", asset.AssetVersion.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error populating asset input for create",
			"Could not create input for creating asset: "+err.Error(),
		)
		return
	}

	createdAsset, err := r.p.Client.CreateAsset(
		asset.OrganizationId.Value,
		asset.EnvironmentId.Value,
		*assetInput,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating asset",
			"Could not create asset, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx, "created asset",
		map[string]interface{}{
			"id":     createdAsset.Id,
			"status": createdAsset.Status,
		},
	)

	result, err := GenerateResourceFromAssetOutput(createdAsset)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating asset",
			"Error when creating asset"+asset.Id.Value+": "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.p.Utils.WaitForAssetStatusInOperationCompleteState(
		ctx,
		result.OrganizationId.Value,
		result.EnvironmentId.Value,
		result.Id.Value,
	); err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for asset on create",
			"Error when waiting for asset id"+result.Id.Value+": "+err.Error(),
		)
		return
	}
}

// Read resource information
func (r resourceAsset) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state Simple
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	assetClientOutput, err := r.p.Client.DescribeAsset(state.OrganizationId.Value, state.EnvironmentId.Value, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading asset",
			"Could not read ID "+state.Id.Value+": "+err.Error(),
		)
		return
	}

	asset, err := GenerateResourceFromAssetOutput(assetClientOutput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error get asset when trying to update (refreshing state)",
			"Could get asset when trying to update (refreshing state): "+state.Id.Value+": "+err.Error(),
		)
		return
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
	var plan Simple
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state and compare against remote
	assetInCloudApi, err := r.p.Client.DescribeAsset(plan.OrganizationId.Value, plan.EnvironmentId.Value, plan.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error update asset",
			"Could not update asset id "+assetInCloudApi.Id+": "+err.Error(),
		)
		return
	}

	rawData, err := json.Marshal(assetInCloudApi)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling vpc plan",
			"Could not marshal vpc plan, unexpected error: "+err.Error(),
		)
		return
	}

	assetInput, err := client.PopulateClientAssetInputForUpdate(ctx, assetInCloudApi, rawData, "vpc", "aws", plan.AssetVersion.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error hydrating asset for update",
			"Error hydrating asset for update - id "+assetInCloudApi.Id+": "+err.Error(),
		)
		return
	}

	// request update
	result, err := r.p.Client.UpdateAsset(
		assetInCloudApi.Id,
		assetInCloudApi.Environment.Id,
		assetInCloudApi.Environment.Organization.Id,
		*assetInput,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error requesting update from cloud api",
			"Could not marshal asset parameters json, unexpected error: "+err.Error(),
		)
		return
	}

	stateToSet, err := GenerateResourceFromAssetOutput(result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error get asset when trying to update (refreshing state)",
			"Could get asset when trying to update (refreshing state): "+assetInCloudApi.Id+": "+err.Error())
		return
	}

	// Set state
	diags = resp.State.Set(ctx, *stateToSet)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.p.Utils.WaitForAssetStatusInOperationCompleteState(
		ctx,
		result.Environment.Organization.Id,
		result.Environment.Id,
		result.Id,
	); err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for asset on update",
			"Error when waiting for asset id"+result.Id+": "+err.Error(),
		)
		return
	}
}

// Delete resource
func (r resourceAsset) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state Simple
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete asset by calling API
	err := r.p.Client.DestroyAsset(state.OrganizationId.Value, state.EnvironmentId.Value, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting asset",
			"Could not delete asset id "+state.Id.Value+": "+err.Error(),
		)
		return
	}

	if err := r.p.Utils.WaitForAssetStatusInOperationCompleteState(
		ctx,
		state.OrganizationId.Value,
		state.EnvironmentId.Value,
		state.Id.Value,
	); err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for asset on delete",
			"Error when waiting for asset id"+state.Id.Value+": "+err.Error(),
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
