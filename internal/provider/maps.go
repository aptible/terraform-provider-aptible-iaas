package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/assets/aws/vpc"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/assets/null/simple"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/environment"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/organization"
)

var (
	// TODO - use freeze or immutable libraries for below maps
	DataSourcesMap = map[string]tfsdk.DataSourceType{
		"aptible_organization": organization.DataSourceOrgType{},
		"aptible_environment":  environment.DataSourceEnvType{},
	}
	ResourcesMap = map[string]tfsdk.ResourceType{
		// aws resources
		"aptible_aws_vpc": vpc.ResourceAssetType{},
		// null resources
		"aptible_null_simple": simple.ResourceAssetType{},
	}
)
