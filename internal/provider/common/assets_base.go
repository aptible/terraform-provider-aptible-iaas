package common

import "github.com/hashicorp/terraform-plugin-framework/types"

type AssetBase struct {
	Id types.String `json:"id" tfsdk:"id"`

	AssetVersion types.String `tfsdk:"asset_version"`
	// Tags map string/string
	// Name
	// Description

	EnvironmentId  types.String `tfsdk:"environment_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Status         types.String `tfsdk:"status"`
}
