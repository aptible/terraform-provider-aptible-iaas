package aptible

import (
	"context"
	"fmt"
	"github.com/aptible/terraform-provider-aptible-iaas/aptible/models"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceOrgsType struct{}

func (r dataSourceOrgsType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"organizations": {
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

func (r dataSourceOrgsType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceOrgs{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceOrgs struct {
	p provider
}

func (r dataSourceOrgs) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Declare struct that this function will set to this data source's state
	var resourceState struct {
		Orgs []models.Org `tfsdk:"orgs"`
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
		resourceState.Orgs = append(resourceState.Orgs, models.Org{
			ID:   org.Id,
			Name: org.Name,
		})
	}

	// Sample debug message
	// To view this message, set the TF_LOG environment variable to DEBUG
	// 		`export TF_LOG=DEBUG`
	// To hide debug message, unset the environment variable
	// 		`unset TF_LOG`
	fmt.Fprintf(stderr, "[DEBUG]-Resource State:%+v", resourceState)

	// Set state
	diags := resp.State.Set(ctx, &resourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
