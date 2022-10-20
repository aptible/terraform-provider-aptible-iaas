package vpc

import (
	"context"
	"fmt"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &VPCDataSource{}

func NewDataSource() datasource.DataSource {
	return &VPCDataSource{}
}

type VPCDataSource struct {
	client client.CloudClient
}

func (r VPCDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Required: true,
			},
			"environment_id": {
				Type:     types.StringType,
				Required: true,
			},
			"organization_id": {
				Type:     types.StringType,
				Required: true,
			},
			"name": {
				Type:     types.StringType,
				Computed: true,
			},
			"status": {
				Type:     types.StringType,
				Computed: true,
			},
			"asset_version": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (d *VPCDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_vpc"
}

func (r *VPCDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(client.CloudClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected client.CloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *VPCDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	vpc, err := r.client.DescribeAsset(ctx, config.OrganizationId.Value, config.EnvironmentId.Value, config.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving org",
			err.Error(),
		)
		return
	}

	state := &ResourceModel{
		Id:             types.String{Value: vpc.Id},
		AssetVersion:   types.String{Value: vpc.AssetVersion},
		EnvironmentId:  types.String{Value: vpc.Environment.Id},
		OrganizationId: types.String{Value: vpc.Environment.Organization.Id},
		Status:         types.String{Value: string(vpc.Status)},
		Name:           types.String{Value: vpc.CurrentAssetParameters.Data["name"].(string)},
	}

	tflog.Info(ctx, "Setting state for asset", map[string]interface{}{"state": state})

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
