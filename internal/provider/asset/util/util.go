package assetutil

import (
	"context"
	"fmt"
	"strings"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
)

var DefaultAssetVersion = "latest"

func extractValues(input []string) (string, string, string) { return input[0], input[1], input[2] }

func StateImporter(ctx context.Context, client client.CloudClient, req resource.ImportStateRequest, resp *resource.ImportStateResponse) *cac.AssetOutput {
	// https://developer.hashicorp.com/terraform/plugin/framework/resources/import#multiple-attributes
	// always found in the following format: "{organization_id},{environment_id},{asset_id}" for this request
	positionalKeys := []string{"organization_id", "environment_id", "asset_id"}
	requestDelimitedValues := strings.Split(req.ID, ",")
	if len(requestDelimitedValues) != 3 {
		resp.Diagnostics.AddError(
			"Error insufficient values to import state",
			fmt.Sprintf("Error unpacking values required for importing state for an asset: Got %d values in csv, expected 3", len(requestDelimitedValues)),
		)
		return nil
	}

	for idx, id := range requestDelimitedValues {
		if _, err := uuid.Parse(id); err != nil {
			resp.Diagnostics.AddError(
				"Error invalid uuid provided to import state",
				fmt.Sprintf("Error in trying to parse uuid (id for %s) from CSV-delimited request: %s",
					positionalKeys[idx], err.Error(),
				),
			)
			return nil
		}
	}

	orgId, envId, assetId := extractValues(requestDelimitedValues)
	assetClientOutput, err := client.DescribeAsset(ctx, orgId, envId, assetId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading asset",
			fmt.Sprintf(
				"Error when reading asset %s: %s",
				req.ID,
				err.Error(),
			),
		)
		return nil
	}

	return assetClientOutput
}
