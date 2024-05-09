package meta

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

type awsAccountKey struct{}

func GetAwsAccountId(ctx context.Context) string {
	x := ctx.Value(awsAccountKey{})
	s, ok := x.(string)
	if ok {
		return s
	}
	return ""
}

func SetAwsAccountId(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, awsAccountKey{}, val)
}

var retryStandard = retry.NewStandard()

func IsErrorRetryable(err error) bool {
	if err == nil {
		return false
	}
	return retryStandard.IsErrorRetryable(err)
}

func ErrorToRequeueResponse(err error) error {
	if IsErrorRetryable(err) {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms())
	}
	return composed.StopWithRequeueDelay(util.Timing.T300000ms())
}

func LogErrorAndReturn(err error, msg string, ctx context.Context) (error, context.Context) {
	result := ErrorToRequeueResponse(err)
	return composed.LogErrorAndReturn(err, msg, result, ctx)
}
