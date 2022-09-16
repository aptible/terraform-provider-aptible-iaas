package vpc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"

	cloud_api_client "github.com/aptible/cloud-api-clients/clients/go"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/common"
)

const DELIMITER = "__"

type ResourceAssetType struct{}

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
}

var vpcAssetParameterSchema = map[string]tfsdk.Attribute{}

func init() {
	maps.Copy(vpcAssetParameterSchema, assetSchema)
}

func (r ResourceAssetType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: vpcAssetParameterSchema,
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

	var asset VPC
	diags := req.Plan.Get(ctx, &asset)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating asset", map[string]interface{}{"asset": asset})
	createdAsset, err := r.p.Client.CreateAsset(
		asset.OrganizationId.String(),
		asset.EnvironmentId.String(),
		cloud_api_client.AssetInput{
			Asset: fmt.Sprintf(
				"aws%svpc%s%s%s",
				DELIMITER,
				DELIMITER,
				asset.AssetVersion,
				DELIMITER,
			),
			AssetVersion: asset.AssetVersion.String(),
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
	tflog.Trace(
		ctx, "created asset",
		map[string]interface{}{
			"id":     createdAsset.Id,
			"status": createdAsset.Status,
		},
	)

	name, _ := createdAsset.CurrentAssetParameters.Data["name"]

	result := VPC{
		AssetBase: common.AssetBase{
			Id:             types.String{Value: createdAsset.Id},
			EnvironmentId:  types.String{Value: createdAsset.Environment.Id},
			OrganizationId: types.String{Value: createdAsset.Environment.Organization.Id},
		},
		AssetParameters: VPCAssetParameters{
			Name: types.String{Value: name.(string)},
		},
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.p.Utils.WaitForAssetStatusInOperationCompleteState(
		ctx,
		result.OrganizationId.String(),
		result.EnvironmentId.String(),
		result.Id.String(),
	); err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for asset on create",
			"Error when waiting for asset id"+result.Id.String()+": "+err.Error(),
		)
		return
	}
}

// Read resource information
func (r resourceAsset) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state VPC
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	assetClientOutput, err := r.p.Client.DescribeAsset(state.OrganizationId.String(), state.EnvironmentId.String(), state.Id.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading asset",
			"Could not read ID "+state.Id.String()+": "+err.Error(),
		)
		return
	}

	name, _ := assetClientOutput.CurrentAssetParameters.Data["name"]

	// interpolate retrieved asset info with existing state
	asset := VPC{
		AssetBase: common.AssetBase{
			Id:             types.String{Value: assetClientOutput.Id},
			EnvironmentId:  types.String{Value: assetClientOutput.Environment.Id},
			OrganizationId: types.String{Value: assetClientOutput.Environment.Organization.Id},
		},

		AssetParameters: VPCAssetParameters{
			Name: types.String{Value: name.(string)},
		},
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
	var plan VPC
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state and compare against remote
	assetInCloudApi, err := r.p.Client.DescribeAsset(plan.OrganizationId.String(), plan.EnvironmentId.String(), plan.Id.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error update asset",
			"Could not update asset id "+assetInCloudApi.Id+": "+err.Error(),
		)
		return
	}

	// request update
	result, err := r.p.Client.UpdateAsset(
		assetInCloudApi.Id,
		assetInCloudApi.Environment.Id,
		assetInCloudApi.Environment.Organization.Id,
		cloud_api_client.AssetInput{
			Asset: fmt.Sprintf(
				"aws%svpc%s%s%s",
				DELIMITER,
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

	name, _ := result.CurrentAssetParameters.Data["name"]

	// Set state
	diags = resp.State.Set(ctx, VPC{
		AssetBase: common.AssetBase{
			Id:             types.String{Value: result.Id},
			EnvironmentId:  types.String{Value: result.Environment.Id},
			OrganizationId: types.String{Value: result.Environment.Organization.Id},
		},

		AssetParameters: VPCAssetParameters{
			Name: types.String{Value: name.(string)},
		},
	})
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
	var state VPC
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete asset by calling API
	err := r.p.Client.DestroyAsset(state.OrganizationId.String(), state.EnvironmentId.String(), state.Id.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting asset",
			"Could not delete asset id "+state.Id.String()+": "+err.Error(),
		)
		return
	}

	if err := r.p.Utils.WaitForAssetStatusInOperationCompleteState(
		ctx,
		state.OrganizationId.String(),
		state.EnvironmentId.String(),
		state.Id.String(),
	); err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for asset on delete",
			"Error when waiting for asset id"+state.Id.String()+": "+err.Error(),
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
