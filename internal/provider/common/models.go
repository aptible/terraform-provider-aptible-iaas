package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

/*
below attributes are common across all assets

	"id": {
		Description: "A valid asset id",
		Type:        types.StringType,
		Computed:    true,
	},
	"environment_id": {
		Description: "A valid environment id",
		Type:        types.StringType,
		Required:    true,
	},
	"organization_id": {
		Description: "A valid organization id",
		Type:        types.StringType,
		Required:    true,
	},
	"asset_platform": {
		Type:     types.StringType,
		Required: true,
	},
	"asset_type": {
		Type:     types.StringType,
		Required: true,
	},
	"asset_version": {
		Type:     types.StringType,
		Required: true,
	},
*/
