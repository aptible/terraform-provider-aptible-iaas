package utils

import "context"

// TODO - move to client? this feels like a client method and there's no need for utils at that point
type UtilsImpl interface {
	WaitForAssetStatusInOperationCompleteState(ctx context.Context, orgId, envId, id string) error
}
