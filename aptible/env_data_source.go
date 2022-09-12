package aptible

import (
	"context"
	"fmt"
	"github.com/aptible/terraform-provider-aptible-iaas/aptible/models"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceEnvType struct{}

func (r dataSourceEnvType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"environments": {
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Type:     types.StringType,
						Computed: true,
					},
				}),
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

func (r dataSourceEnv) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var resourceState struct {
		Env *models.Environment `tfsdk:"env"`
	}

	var config models.Environment
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envs, err := r.p.client.ListEnvironments(config.OrgID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving environments",
			err.Error(),
		)
		return
	}

	for _, env := range envs {
		if env.Id == config.ID {
			resourceState.Env = &models.Environment{
				ID:    env.Id,
				Name:  env.Name,
				OrgID: config.OrgID, // TODO: Can I do: env.Organization.Id?
			}
			break
		}
	}
	if resourceState.Env == nil {
		resp.Diagnostics.AddError(
			"Could not find environment",
			config.ID,
		)
	}

	// To view this message, set the TF_LOG environment variable to DEBUG
	// 		`export TF_LOG=DEBUG`
	fmt.Fprintf(stderr, "[DEBUG]-Resource State:%+v", resourceState)

	// Set state
	diags = resp.State.Set(ctx, &resourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
