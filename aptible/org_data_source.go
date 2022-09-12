package aptible

import (
	"context"
	"fmt"
	"github.com/aptible/terraform-provider-aptible-iaas/aptible/models"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceOrgType struct{}

func (r dataSourceOrgType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Required: true,
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

func (r dataSourceOrg) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var resourceState struct {
		Org *models.Org `tfsdk:"org"`
	}

	var config models.Org
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgs, err := r.p.client.ListOrgs()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving coffee",
			err.Error(),
		)
		return
	}

	for _, org := range orgs {
		if org.Id == config.ID {
			resourceState.Org = &models.Org{
				ID:   org.Id,
				Name: org.Name,
			}
			break
		}
	}

	if resourceState.Org == nil {
		resp.Diagnostics.AddError(
			"Could not find organization",
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
