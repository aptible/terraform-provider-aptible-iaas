package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type AssetBase struct {
	Id types.String `json:"id" tfsdk:"id"`

	AssetPlatform types.String `tfsdk:"asset_platform"`
	AssetType     types.String `tfsdk:"asset_type"`
	AssetVersion  types.String `tfsdk:"asset_version"`

	EnvironmentId  types.String `tfsdk:"environment_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Status         types.String `tfsdk:"status"`
}

type Asset struct {
	AssetBase
	Parameters types.String `tfsdk:"parameters"`
}
