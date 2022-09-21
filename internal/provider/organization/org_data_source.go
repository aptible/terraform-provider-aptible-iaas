package organization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/common"
)

type DataSourceOrgType struct{}

func (r DataSourceOrgType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (r DataSourceOrgType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceOrg{
		p: *(p.(*common.Provider)),
	}, nil
}

type dataSourceOrg struct {
	p common.Provider
}

func (r dataSourceOrg) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var config Org
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.p.Client.FindOrg(config.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving org",
			err.Error(),
		)
		return
	}

	state := &Org{
		ID:   types.String{Value: org.Id},
		Name: types.String{Value: org.Name},
	}

	tflog.Debug(ctx, "Creating asset", map[string]interface{}{"state": state})

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
