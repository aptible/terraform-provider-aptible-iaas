package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/provider"
)

func main() {
	providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "aptible.com/aptible/aptible-iaas",
	})
}
