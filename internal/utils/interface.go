package utils

import "context"

type UtilsImpl interface {
	WaitForAssetStatusInOperationCompleteState(ctx context.Context, orgId, envId, id string) error
}
