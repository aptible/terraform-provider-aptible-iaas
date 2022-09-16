package utils

import (
	"context"
	"fmt"
	"time"

	cloud_api_client "github.com/aptible/cloud-api-clients/clients/go"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// DefaultTimeToWait - typical time.sleep to wait before waiting to try agian
var DefaultTimeToWait = 10 * time.Second

// TimeToFail - maximum time to wait before failing a given operation completion
var TimeToFail = 30 * time.Minute

// AssetStatusesThatIndicateCompletion - these statuses indicate the asset requested is now in a somewhat
// final state and is ready for operations
var AssetStatusesThatIndicateCompletion = []cloud_api_client.AssetStatus{
	cloud_api_client.ASSETSTATUS_FAILED,
	cloud_api_client.ASSETSTATUS_DEPLOYED,
	cloud_api_client.ASSETSTATUS_DESTROYED,
}

// ErrorTimeOutOnAssetStatus - error that's returned when asset waiter times out
var ErrorTimeOutOnAssetStatus = fmt.Errorf("timed out when waiting for asset status")

func (u Utils) WaitForAssetStatusInOperationCompleteState(ctx context.Context, orgId, envId, id string) error {
	tflog.Trace(
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
			return ErrorTimeOutOnAssetStatus
		}
		asset, err := u.client.DescribeAsset(orgId, envId, id)
		if err != nil {
			return err
		}
		for _, completedOperationStatus := range AssetStatusesThatIndicateCompletion {
			if completedOperationStatus == asset.Status {
				tflog.Info(ctx, "Completed waiting for operation, in state ready for further operations")
				// if operation is completed, just break, we're done!
				return nil
			}
		}
		// actively wait for status
		time.Sleep(DefaultTimeToWait)
		tflog.Trace(ctx,
			"Still waiting for status on provided asset",
			map[string]interface{}{
				"id":               id,
				"envId":            envId,
				"orgId":            orgId,
				"totalTimeRunning": totalTimeRunning,
			})
	}
}
