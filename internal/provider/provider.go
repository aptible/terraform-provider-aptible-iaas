package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/common"
)

func New() tfsdk.Provider {
	return &common.Provider{
		ResourcesMap:   ResourcesMap,
		DataSourcesMap: DataSourcesMap,
	}
}
