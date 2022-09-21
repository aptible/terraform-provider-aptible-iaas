package connection

import "github.com/hashicorp/terraform-plugin-framework/types"

type Connection struct {
	// TODO
	Name types.String `tfsdk:"name"`
	// TODO
	Description    types.String `tfsdk:"description"`
	Id             types.String `tfsdk:"id"`
	EnvironmentId  types.String `tfsdk:"environment_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	// AssetId - note this corresponds to the INCOMING connection asset id, this is mapped
	// in this way because in our RESTful urls, the `asset_id` corresponds to
	// the incoming_asset_connection
	AssetId                 types.String `tfsdk:"incoming_asset_connection"`
	OutgoingAssetConnection types.String `tfsdk:"outgoing_asset_connection"`
}
