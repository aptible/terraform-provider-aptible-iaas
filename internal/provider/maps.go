package provider

import (
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/assets/aws/vpc"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/environment"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/organization"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	// TODO - use freeze or immutable libraries for below maps
	DataSourcesMap = map[string]tfsdk.DataSourceType{
		"aptible_organization": organization.DataSourceOrgType{},
		"aptible_environment":  environment.DataSourceEnvType{},
	}
	ResourcesMap = map[string]tfsdk.ResourceType{
		"aptible_aws_vpc": vpc.ResourceAssetType{},
	}
)
