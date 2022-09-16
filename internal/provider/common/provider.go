package common

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/utils"
	"os"
)

type Provider struct {
	ProviderRootContext context.Context
	Configured          bool
	Client              client.CloudClient
	Utils               utils.UtilsImpl
}

// GetSchema ...
func (p *Provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

// providerData schema struct
type providerData struct {
	Token types.String `tfsdk:"token"`
	Host  types.String `tfsdk:"host"`
	Debug types.Bool   `tfsdk:"debug"`
}

func (p *Provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	// Retrieve Provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// User must provide a user to the Provider
	var token string
	if config.Token.Unknown {
		// Cannot connect to Client with an unknown value
		resp.Diagnostics.AddWarning(
			"Unable to create Client",
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
		// Cannot connect to Client with an unknown value
		resp.Diagnostics.AddError(
			"Unable to create Client",
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
		// Cannot connect to Client with an unknown value
		resp.Diagnostics.AddError(
			"Unable to create Client",
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

	p.Client = c
	p.Utils = utils.NewUtils(p.Client)
	p.Configured = true
}

// GetResources - Defines Provider resources
func (p *Provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return ctx.Value("aptible_resources").(map[string]tfsdk.ResourceType), nil
}

// GetDataSources - Defines Provider data sources
func (p *Provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return ctx.Value("aptible_data_sources").(map[string]tfsdk.DataSourceType), nil
}
