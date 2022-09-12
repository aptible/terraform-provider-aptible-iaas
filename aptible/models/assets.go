package models

type AssetBase struct {
	Id string `json:"id" tfsdk:"id"`

	AssetPlatform string `tfsdk:"asset_platform"`
	AssetType     string `tfsdk:"asset_type"`
	AssetVersion  string `tfsdk:"asset_version"`

	EnvironmentId  string `tfsdk:"environment_id"`
	OrganizationId string `tfsdk:"organization_id"`
	Status         string `tfsdk:"status"`
}

type Asset struct {
	AssetBase
	Parameters string `tfsdk:"parameters"`
}
