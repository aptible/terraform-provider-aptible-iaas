package aptible

import (
	"context"
	"os"

	"github.com/aptible/terraform-provider-aptible-iaas/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var stderr = os.Stderr

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     client.CloudClient
}

// GetSchema
func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"token": {
				Type:     types.StringType,
				Required: true,
			},
			"host": {
				Type:     types.StringType,
				Optional: true,
			},
			"debug": {
				Type:     types.BoolType,
				Optional: true,
			},
		},
	}, nil
}

// Provider schema struct
type providerData struct {
	Token types.String `tfsdk:"token"`
	Host  types.String `tfsdk:"host"`
	Debug types.Bool   `tfsdk:"debug"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// User must provide a user to the provider
	var token string
	if config.Token.Unknown {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as username",
		)
		return
	}

	if config.Token.Null {
		token = os.Getenv("APTIBLE_TOKEN")
	} else {
		token = config.Token.Value
	}

	if token == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find username",
			"Username cannot be an empty string",
		)
		return
	}

	var debug bool
	if config.Debug.Unknown {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Cannot use unknown value as debug",
		)
		return
	}

	if config.Debug.Null {
		debugStr := os.Getenv("APTIBLE_DEBUG")
		if debugStr == "1" {
			debug = true
		} else {
			debug = false
		}
	} else {
		debug = config.Debug.Value
	}

	// User must specify a host
	var host string
	if config.Host.Unknown {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Cannot use unknown value as host",
		)
		return
	}

	if config.Host.Null {
		host = os.Getenv("APTIBLE_HOST")
	} else {
		host = config.Host.Value
	}

	if host == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find host",
			"Host cannot be an empty string",
		)
		return
	}

	c := client.NewClient(debug, host, token)

	p.client = c
	p.configured = true
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"asset": resourceAssetType{},
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"aptible_organization": dataSourceOrgType{},
		"aptible_environment":  dataSourceEnvType{},
	}, nil
}
