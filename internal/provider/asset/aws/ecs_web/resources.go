/*
aws/acm/resources.go is the template that we copy and paste to other
aws asset resource files using `make resource`.

ONLY edit aws/acm/resources.go.
*/
package ecsweb

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/util"
)

var _ resource.ResourceWithImportState = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client client.CloudClient
}

func (r Resource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: resourceDescription,
		Attributes:          AssetSchema,
	}, nil
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + resourceTypeName
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(client.CloudClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.CloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating asset", map[string]interface{}{"asset": plan})

	assetInput, err := planToAssetInput(ctx, plan)
	if err != nil {
		return
	}

	createdAsset, err := r.client.CreateAsset(
		ctx,
		plan.OrganizationId.Value,
		plan.EnvironmentId.Value,
		assetInput,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating asset",
			"Could not create asset, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Info(
		ctx, "created asset",
		map[string]interface{}{
			"id":     createdAsset.Id,
			"status": createdAsset.Status,
		},
	)

	nextPlan, err := assetOutputToPlan(ctx, plan, createdAsset)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating asset",
			fmt.Sprintf(
				"Error when creating asset %s: %s",
				plan.Id.Value,
				err.Error(),
			),
		)
		return
	}

	diags = resp.State.Set(ctx, nextPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	completedAsset, err := util.WaitForAssetStatusInOperationCompleteState(
		r.client,
		ctx,
		plan.OrganizationId.Value,
		plan.EnvironmentId.Value,
		createdAsset.Id,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for asset on create",
			fmt.Sprintf(
				"Error when waiting for asset id %s: %s",
				createdAsset.Id,
				err.Error(),
			),
		)
		return
	}

	nextPlan, err = assetOutputToPlan(ctx, plan, completedAsset)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating asset",
			fmt.Sprintf(
				"Error when creating asset %s: %s",
				plan.Id.Value,
				err.Error(),
			),
		)
		return
	}

	diags = resp.State.Set(ctx, nextPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	assetClientOutput, err := r.client.DescribeAsset(ctx, state.OrganizationId.Value, state.EnvironmentId.Value, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading asset",
			fmt.Sprintf(
				"Error when creating asset %s: %s",
				state.Id.Value,
				err.Error(),
			),
		)
		return
	}

	asset, err := assetOutputToPlan(ctx, state, assetClientOutput)
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

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan ResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan values
	var state ResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state and compare against remote
	assetInCloudApi, err := r.client.DescribeAsset(
		ctx,
		plan.OrganizationId.Value,
		plan.EnvironmentId.Value,
		state.Id.Value,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error update asset",
			"Could not update asset id "+assetInCloudApi.Id+": "+err.Error(),
		)
		return
	}

	assetInput, err := planToAssetInput(ctx, plan)
	if err != nil {
		return
	}

	// request update
	result, err := r.client.UpdateAsset(
		ctx,
		assetInCloudApi.Id,
		assetInCloudApi.Environment.Id,
		assetInCloudApi.Environment.Organization.Id,
		assetInput,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error requesting update from cloud api",
			fmt.Sprintf("Could not marshal asset parameters json, unexpected error: %s", err.Error()),
		)
		return
	}

	completedAsset, err := util.WaitForAssetStatusInOperationCompleteState(
		r.client,
		ctx,
		result.Environment.Organization.Id,
		result.Environment.Id,
		result.Id,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for asset on update",
			fmt.Sprintf("Error when waiting for asset id: %s: %s", result.Id, err.Error()),
		)
		return
	}

	stateToSet, err := assetOutputToPlan(ctx, plan, completedAsset)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error get asset when trying to update (refreshing state)",
			fmt.Sprintf("Could get asset when trying to update (refreshing state): %s: %s", assetInCloudApi.Id, err.Error()),
		)
		return
	}

	// Set state
	diags = resp.State.Set(ctx, *stateToSet)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete asset by calling API
	err := r.client.DestroyAsset(ctx, state.OrganizationId.Value, state.EnvironmentId.Value, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting asset",
			fmt.Sprintf("Could not delete asset id %s: %s", state.Id.Value, err.Error()),
		)
		return
	}

	_, err = util.WaitForAssetStatusInOperationCompleteState(
		r.client,
		ctx,
		state.OrganizationId.Value,
		state.EnvironmentId.Value,
		state.Id.Value,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for asset on delete",
			fmt.Sprintf("Error when waiting for asset id %s: %s", state.Id.Value, err.Error()),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
