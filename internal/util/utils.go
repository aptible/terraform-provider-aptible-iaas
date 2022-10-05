package util

import (
	"context"
	"fmt"
	"time"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// DefaultTimeToWait - typical time.sleep to wait before waiting to try agian
var DefaultTimeToWait = 10 * time.Second

// TimeToFail - maximum time to wait before failing a given operation completion
var TimeToFail = 30 * time.Minute

// AssetStatusesThatIndicateCompletion - these statuses indicate the asset requested is now in a somewhat
// final state and is ready for operations
var AssetStatusesThatIndicateCompletion = []cac.AssetStatus{
	cac.ASSETSTATUS_DEPLOYED,
	cac.ASSETSTATUS_DESTROYED,
}

// ErrorTimeOutOnAssetStatus - error that's returned when asset waiter times out
var ErrorTimeOutOnAssetStatus = fmt.Errorf("timed out when waiting for asset status")

func WaitForAssetStatusInOperationCompleteState(client client.CloudClient, ctx context.Context, orgId, envId, id string) (*cac.AssetOutput, error) {
	tflog.Info(
		ctx, "waiting for asset status",
		map[string]interface{}{
			"id":    id,
			"envId": envId,
		},
	)
	totalTimeRunning := time.Now().Add(TimeToFail)

	for {
		if totalTimeRunning.Before(time.Now()) {
			tflog.Error(
				ctx,
				"Error when waiting for status",
				map[string]interface{}{
					"id":               id,
					"envId":            envId,
					"orgId":            orgId,
					"totalTimeRunning": totalTimeRunning,
				})
			return nil, ErrorTimeOutOnAssetStatus
		}
		asset, err := client.DescribeAsset(ctx, orgId, envId, id)
		if err != nil {
			return nil, err
		}

		if asset.Status == cac.ASSETSTATUS_FAILED {
			return nil, fmt.Errorf("Asset status FAILED %s", id)
		}

		for _, completedOperationStatus := range AssetStatusesThatIndicateCompletion {
			if completedOperationStatus == asset.Status {
				tflog.Info(ctx, "Completed waiting for operation, in state ready for further operations")
				// if operation is completed, just break, we're done!
				return asset, nil
			}
		}
		// actively wait for status
		time.Sleep(DefaultTimeToWait)
		tflog.Info(ctx,
			"Still waiting for status on provided asset",
			map[string]interface{}{
				"id":               id,
				"envId":            envId,
				"orgId":            orgId,
				"totalTimeRunning": totalTimeRunning,
			})
	}
}

func SafeString(obj interface{}) string {
	switch val := obj.(type) {
	case string:
		return val
	default:
		return ""
	}
}