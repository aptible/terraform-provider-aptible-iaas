package models

type AssetBase struct {
	Id          string `tfsdk:"id"`
	Name        string `tfsdk:"name"`
	Description string `tfsdk:"description"`
	Version     string `tfsdk:"version"`

	AssetPlatform string `tfsdk:"asset_platform"`
	AssetType     string `tfsdk:"asset_type"`
	AssetVersion  string `tfsdk:"asset_version"`

	EnvironmentId  string `tfsdk:"environment_id"`
	OrganizationId string `tfsdk:"organization_id"`
}

type Asset struct {
	AssetBase
	Parameters string `tfsdk:"parameters"`
}
