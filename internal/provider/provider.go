package provider

import (
	"context"
	"encoding/json"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/asset/aws/acm"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/asset/aws/ecs_compute"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/asset/aws/ecs_web"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/asset/aws/rds"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/asset/aws/redis"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/asset/aws/secret"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/asset/aws/vpc"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/environment"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/organization"
)

var _ provider.Provider = &Provider{}

type Provider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "aptible"
	resp.Version = p.version
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
			"auth_host": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

// providerData schema struct
type providerData struct {
	Token    types.String `tfsdk:"token"`
	AuthHost types.String `tfsdk:"auth_host"`
	Host     types.String `tfsdk:"host"`
}

func extractValueFromTokensJson(config *providerData) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// skip over this, silently?
		return config.Token.Value
	}
	tokensJson, err := os.ReadFile(homeDir + "/.aptible/tokens.json")
	if err != nil {
		return config.Token.Value
	}

	var output map[string]interface{}
	if err = json.Unmarshal(tokensJson, &output); err != nil {
		return config.Token.Value
	}

	// find if in host and specified
	if !config.AuthHost.Null {
		if _, found := output[config.AuthHost.Value]; found {
			return output[config.AuthHost.Value].(string)
		}
	}

	// fall back to default host if no host specified
	if _, found := output["https://auth.aptible.com"]; found {
		return output["https://auth.aptible.com"].(string)
	}

	return config.Token.Value
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
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

	// APTIBLE_TOKEN retrieval
	// 1 - try to use the token passed in on the provider stanza
	// 2 - by default, try to use the environment variable
	// 3 - if no environment variable specified, fall back to API token
	if config.Token.Null {
		token = os.Getenv("APTIBLE_TOKEN")
		if token == "" {
			token = extractValueFromTokensJson(&config)
		}
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

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		acm.NewResource,
		ecsweb.NewResource,
		rds.NewResource,
		redis.NewResource,
		vpc.NewResource,
		secret.NewResource,
		ecscompute.NewResource,
	}
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		organization.NewDataSource,
		environment.NewDataSource,
		vpc.NewDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{
			version: version,
		}
	}
}
