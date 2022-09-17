package provider

import (
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/assets/aws/vpc"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/environment"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/organization"
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
