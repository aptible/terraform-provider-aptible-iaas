package aptible

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/aptible/terraform-provider-aptible-iaas/aptible/models"
)

type dataSourceEnvType struct{}

func (r dataSourceEnvType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Required: true,
			},
			"org_id": {
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

func (r dataSourceEnvType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceEnv{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceEnv struct {
	p provider
}

type envConfig struct {
	ID    types.String `tfsdk:"id"`
	OrgId types.String `tfsdk:"org_id"`
	Name  types.String `tfsdk:"name"`
}

func (r dataSourceEnv) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var config envConfig
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.p.client.DescribeEnvironment(config.OrgId.String(), config.ID.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving environments",
			err.Error(),
		)
		return
	}

	state := &models.Environment{
		ID:    env.Id,
		OrgID: env.Organization.Id,
		Name:  env.Name,
	}

	tflog.Debug(ctx, "Creating asset", map[string]interface{}{"state": state})

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
