package aptible

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/aptible/terraform-provider-aptible-iaas/aptible/models"
)

type dataSourceOrgType struct{}

func (r dataSourceOrgType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Required: true,
			},
			"name": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (r dataSourceOrgType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceOrg{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceOrg struct {
	p provider
}

type orgConfig struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (r dataSourceOrg) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var config orgConfig
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.p.client.FindOrg(config.ID.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving org",
			err.Error(),
		)
		return
	}

	state := &models.Org{
		ID:   org.Id,
		Name: org.Name,
	}

	tflog.Debug(ctx, "Creating asset", map[string]interface{}{"state": state})

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
