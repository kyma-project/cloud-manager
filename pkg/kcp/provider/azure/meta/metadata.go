package meta

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func TooManyRequests(err error) bool {
	var respErr *azcore.ResponseError

	// https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/request-limits-and-throttling
	return errors.As(err, &respErr) && respErr.StatusCode == 429
}

func ErrorToRequeueResponse(err error) error {
	if TooManyRequests(err) {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms())
	}
	return composed.StopWithRequeueDelay(util.Timing.T300000ms())
}

func LogErrorAndReturn(err error, msg string, ctx context.Context) (error, context.Context) {
	result := ErrorToRequeueResponse(err)
	return composed.LogErrorAndReturn(err, msg, result, ctx)
}
