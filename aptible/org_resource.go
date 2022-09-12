package aptible

import (
	"context"
	"math/big"
	"strconv"
	"time"

	cloud_api_client "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type resourceOrgType struct{}

// Order Resource schema
func (r resourceOrgType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"organization_id": {
				Description: "A valid organization id",
				Type:        types.StringType,
				Required:    true,
			},
			"baa_status": {
				Type:     types.StringType,
				Required: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"aws_account_ou": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

// New resource instance
func (r resourceOrgType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceOrg{
		p: *(p.(*provider)),
	}, nil
}

type resourceOrg struct {
	p provider
}

// Create a new resource
func (r resourceOrg) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var org Org
	diags := req.Plan.Get(ctx, &org)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating organization")

	o, err := r.p.client.CreateOrg(org.ID, cloud_api_client.OrganizationInput{
		Name: org.Name,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}
	// for more information on logging from providers, refer to
	// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog
	tflog.Trace(ctx, "created org", map[string]interface{}{"id": o.Id, "ou": o.AwsOu})

	result := Org{
		ID:   o.Id,
		Name: o.Name,
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r resourceOrg) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state Org
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get order from API and then update what is in state from what the API returns
	orgID := state.ID

	// Get order current value
	org, err := r.p.client.FindOrg(orgID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading order",
			"Could not read orderID "+orgID+": "+err.Error(),
		)
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &org)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r resourceOrder) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan values
	var plan Org
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state Org
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get order ID from state
	orgID := state.ID

	// Update order by calling API
	order, err := r.p.client
	if err != nil {
		resp.Diagnostics.AddError(
			"Error update order",
			"Could not update orderID "+orderID+": "+err.Error(),
		)
		return
	}

	// Map response body to resource schema attribute
	var ois []OrderItem
	for _, oi := range order.Items {
		ois = append(ois, OrderItem{
			Coffee: Coffee{
				ID:          oi.Coffee.ID,
				Name:        types.String{Value: oi.Coffee.Name},
				Teaser:      types.String{Value: oi.Coffee.Teaser},
				Description: types.String{Value: oi.Coffee.Description},
				Price:       types.Number{Value: big.NewFloat(oi.Coffee.Price)},
				Image:       types.String{Value: oi.Coffee.Image},
			},
			Quantity: oi.Quantity,
		})
	}

	// Generate resource state struct
	var result = Order{
		ID:          types.String{Value: strconv.Itoa(order.ID)},
		Items:       ois,
		LastUpdated: types.String{Value: string(time.Now().Format(time.RFC850))},
	}

	// Set state
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceOrder) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state Order
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get order ID from state
	orderID := state.ID.Value

	// Delete order by calling API
	err := r.p.client.DeleteOrder(orderID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting order",
			"Could not delete orderID "+orderID+": "+err.Error(),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// Import resource
func (r resourceOrder) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
