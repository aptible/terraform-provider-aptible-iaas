package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/assets/vpc"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/environment"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/organization"
)

const (
	DATA_SOURCES_CONTEXT_KEY = "aptible_data_sources"
	RESOURCES_CONTEXT_KEY    = "aptible_resources"
)

var (
	DataSourcesMap = map[string]tfsdk.DataSourceType{
		"aptible_organization": organization.DataSourceOrgType{},
		"aptible_environment":  environment.DataSourceEnvType{},
	}
	ResourcesMap = map[string]tfsdk.ResourceType{
		"aptible_aws_vpc": vpc.ResourceAssetType{},
	}
)
