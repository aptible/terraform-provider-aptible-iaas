package main

import (
	"context"

	"github.com/aptible/terraform-provider-aptible-iaas/aptible"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	providerserver.Serve(context.Background(), aptible.New, providerserver.ServeOpts{
		Address: "aptible.com/aptible/aptible-iaas",
	})
}
