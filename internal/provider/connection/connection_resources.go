package connection

import (
	"context"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type connectionType struct{}

var connectionSchema = map[string]tfsdk.Attribute{
	"id": {
		Description: "A connection id",
		Type:        types.StringType,
		Computed:    true,
	},
	"name": {
		Description: "A name that can be given to a connection",
		Type:        types.StringType,
		Optional:    true,
	},
	"description": {
		Description: "A description that can be given to a connection",
		Type:        types.StringType,
		Optional:    true,
	},
	"incoming_asset_connection": {
		Description: "An incoming asset id for a connection",
		Type:        types.StringType,
		Required:    true,
	},
	"outgoing_asset_connection": {
		Description: "An outgoing asset id for a connection",
		Type:        types.StringType,
		Required:    true,
	},
}

func (r connectionType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: connectionSchema,
	}, nil
}

func (r connectionType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceConnection{
		p: *(p.(*provider.provider)),
	}, nil
}

type resourceConnection struct {
	p provider.provider
}

func (r resourceConnection) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	panic("not implemented!")
}

// Read resource information
func (r resourceConnection) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state Connection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionClientOutput, err := r.p.client.GetConnection(
		state.OrganizationId.Value,
		state.EnvironmentId.Value,
		state.AssetId.Value,
		state.Id.Value,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading connection",
			"Could not read ID "+state.Id.Value+": "+err.Error(),
		)
		return
	}

	// TODO FINISH THIS
	_ = Connection{
		Name:                    state.Name,        // TODO, currently not assigned to backend or sent
		Description:             state.Description, // TODO, currently not assigned to backend or sent
		Id:                      types.String{Value: connectionClientOutput.Id},
		EnvironmentId:           types.String{Value: connectionClientOutput.IncomingConnectionAsset.Environment.Id},
		OrganizationId:          types.String{Value: connectionClientOutput.IncomingConnectionAsset.Environment.Organization.Id},
		AssetId:                 types.String{Value: connectionClientOutput.IncomingConnectionAsset.Id},
		OutgoingAssetConnection: types.String{Value: connectionClientOutput.OutgoingConnectionAsset.Id},
	}
}

func (r resourceConnection) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	panic("not implemented!")
}

// Delete resource
func (r resourceConnection) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	panic("not implemented!")
}

func (r resourceConnection) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
