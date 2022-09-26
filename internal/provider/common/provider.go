package common

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/utils"
)

type Provider struct {
	ResourcesMap   map[string]tfsdk.ResourceType
	DataSourcesMap map[string]tfsdk.DataSourceType
	Configured     bool
	Client         client.CloudClient
	Utils          utils.UtilsImpl
}

// GetSchema ...
func (p *Provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"token": {
				Type:     types.StringType,
				Optional: true,
			},
			"host": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

// providerData schema struct
type providerData struct {
	Token types.String `tfsdk:"token"`
	Host  types.String `tfsdk:"host"`
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
			"Cannot use unknown value as token",
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
			"Unable to find token",
			"Token cannot be an empty string",
		)
		return
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

	c := client.NewClient(true, host, token)

	p.Client = c
	p.Utils = utils.NewUtils(p.Client)
	p.Configured = true
}

// GetResources - Defines Provider resources
func (p *Provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return p.ResourcesMap, nil
}

// GetDataSources - Defines Provider data sources
func (p *Provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return p.DataSourcesMap, nil
}
