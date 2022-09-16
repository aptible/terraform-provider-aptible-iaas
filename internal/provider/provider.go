package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider/common"
)

func New() tfsdk.Provider {
	ctx := context.Background()
	ctx = context.WithValue(ctx, RESOURCES_CONTEXT_KEY, ResourcesMap)
	ctx = context.WithValue(ctx, DATA_SOURCES_CONTEXT_KEY, DataSourcesMap)
	return &common.Provider{
		ProviderRootContext: ctx,
	}
}
