package utils

import (
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
)

type Utils struct {
	client client.CloudClient
}

func NewUtils(client client.CloudClient) UtilsImpl {
	return Utils{
		client: client,
	}
}
