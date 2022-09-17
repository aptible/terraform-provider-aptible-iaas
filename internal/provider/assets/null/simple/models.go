package simple

import (
	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SimpleAssetParameters struct {
	Name types.String
}

type Simple struct {
	common.AssetBase
	AssetParameters SimpleAssetParameters
}
