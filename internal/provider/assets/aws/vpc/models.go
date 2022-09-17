package vpc

import (
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TODO - dynamically generate this
type VPCAssetParameters struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

type VPC struct {
	common.AssetBase
	AssetParameters VPCAssetParameters
}
