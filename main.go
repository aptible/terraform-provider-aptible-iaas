package main

import (
	"context"
	"terraform-provider-hashicups-pf/hashicups"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	providerserver.Serve(context.Background(), hashicups.New, providerserver.ServeOpts{
		Address: "hashicorp.com/aptible-iaas",
	})
}
