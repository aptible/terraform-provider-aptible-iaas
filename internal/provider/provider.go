package provider

import (
	"context"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/asset/aws/s3"
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
		},
	}, nil
}

// providerData schema struct
type providerData struct {
	Token types.String `tfsdk:"token"`
	Host  types.String `tfsdk:"host"`
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
		s3.NewResource,
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
