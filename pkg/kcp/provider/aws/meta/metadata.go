package meta

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/retry"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

const (
	UnauthorizedOperation = "UnauthorizedOperation"
	AccessDenied          = "AccessDenied"
	RouteNotSupported     = "RouteNotSupported"
)

type awsAccountKeyType struct{}

var awsAccountKey = awsAccountKeyType{}

func GetAwsAccountId(ctx context.Context) string {
	x := ctx.Value(awsAccountKey)
	s, ok := x.(string)
	if ok {
		return s
	}
	return ""
}

func SetAwsAccountId(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, awsAccountKey, val)
}

var retryStandard = retry.NewStandard()

func IsErrorRetryable(err error) bool {
	if err == nil {
		return false
	}
	return retryStandard.IsErrorRetryable(err)
}

func AsApiError(err error) smithy.APIError {
	if apiError, ok := errors.AsType[smithy.APIError](err); ok {
		return apiError
	}
	return nil
}

func GetErrorMessage(err error, def string) (string, bool) {
	if apiError, ok := errors.AsType[smithy.APIError](err); ok {
		switch apiError.ErrorCode() {
		case UnauthorizedOperation:
			return "Not authorized to perform this operation.", true
		case AccessDenied:
			return "Not authorized to assume role.", true
		case RouteNotSupported:
			return "Route not supported.", true
		}
	}

	return def, false
}

func IsAccessDenied(err error) bool {
	if apiError, ok := errors.AsType[smithy.APIError](err); ok {
		return apiError.ErrorCode() == AccessDenied
	}
	return false
}

func IsUnauthorized(err error) bool {
	if apiError, ok := errors.AsType[smithy.APIError](err); ok {
		return apiError.ErrorCode() == UnauthorizedOperation
	}
	return false
}

func IsRouteNotSupported(err error) bool {
	if apiError, ok := errors.AsType[smithy.APIError](err); ok {
		return apiError.ErrorCode() == RouteNotSupported
	}
	return false
}

// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/errors-overview.html
var notFoundErrorCodes = map[string]struct{}{
	(&acmtypes.ResourceNotFoundException{}).ErrorCode():             {},
	(&efstypes.FileSystemNotFound{}).ErrorCode():                    {},
	(&efstypes.AccessPointNotFound{}).ErrorCode():                   {},
	(&efstypes.MountTargetNotFound{}).ErrorCode():                   {},
	(&efstypes.PolicyNotFound{}).ErrorCode():                        {},
	(&elasticachetypes.CacheSubnetGroupNotFoundFault{}).ErrorCode(): {},
	(&elasticachetypes.CacheClusterNotFoundFault{}).ErrorCode():     {},
	(&secretsmanagertypes.ResourceNotFoundException{}).ErrorCode():  {},
	(&wafv2types.WAFNonexistentItemException{}).ErrorCode():         {},
	"InvalidVpcPeeringConnectionID.NotFound":                        {},
	"InvalidVpcID.NotFound":                                         {},
	"InvalidRoute.NotFound":                                         {},
}

func IsNotFound(err error) bool {
	if err != nil {
		if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
			_, listed := notFoundErrorCodes[apiErr.ErrorCode()]
			if listed {
				return true
			}
		}

		if respErr, ok := errors.AsType[*smithyhttp.ResponseError](err); ok {
			return respErr.HTTPStatusCode() == http.StatusNotFound
		}
	}
	return false
}

func NewHttpNotFoundError(err error) error {
	return &smithyhttp.ResponseError{
		Err: err,
		Response: &smithyhttp.Response{
			Response: &http.Response{
				StatusCode: 404,
			},
		},
	}
}
func RetryableErrorToRequeueResponse(err error) error {
	if IsErrorRetryable(err) {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms())
	}
	return nil
}

func ErrorToRequeueResponse(err error) error {
	if err == nil {
		return nil
	}
	if IsErrorRetryable(err) {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms())
	}
	return composed.StopWithRequeue
}

func LogErrorAndReturn(err error, msg string, ctx context.Context) (error, context.Context) {
	result := ErrorToRequeueResponse(err)
	return composed.LogErrorAndReturn(err, msg, result, ctx)
}

type ElastiCacheState = string

// github.com/aws/aws-sdk-go-v2/service/elasticache@v1.40.3/types/types.go
// Status *string
// The current state of this replication group - creating , available , modifying ,
// deleting , create-failed , snapshotting .
const (
	ElastiCache_AVAILABLE     ElastiCacheState = "available"
	ElastiCache_CREATING      ElastiCacheState = "creating"
	ElastiCache_DELETING      ElastiCacheState = "deleting"
	ElastiCache_MODIFYING     ElastiCacheState = "modifying"
	ElastiCache_CREATE_FAILED ElastiCacheState = "create-failed"
	ElastiCache_SNAPSHOTTING  ElastiCacheState = "snapshotting"
)

type ElastiCacheUserGroupState = string

const (
	ElastiCache_UserGroup_ACTIVE    ElastiCacheUserGroupState = "active"
	ElastiCache_UserGroup_CREATING  ElastiCacheUserGroupState = "creating"
	ElastiCache_UserGroup_DELETING  ElastiCacheUserGroupState = "deleting"
	ElastiCache_UserGroup_MODIFYING ElastiCacheUserGroupState = "modifying"
)

// ElastiCacheProvisioningTimeout is the default maximum time allowed for an ElastiCache provisioning
// operation. It is a package-level var (not a const) to allow envtests to inject a shorter deadline;
// no production code writes it. The default of 2h is the intended operational timeout.
var ElastiCacheProvisioningTimeout = 2 * time.Hour

// ProvisioningAction is the decision returned by EvaluateProvisioningAction.
type ProvisioningAction int

const (
	ProvisioningActionWait    ProvisioningAction = iota // keep waiting
	ProvisioningActionTimeout                           // deadline exceeded
	ProvisioningActionHeal                              // trigger self-heal
)

func (a ProvisioningAction) String() string {
	switch a {
	case ProvisioningActionWait:
		return "Wait"
	case ProvisioningActionTimeout:
		return "Timeout"
	case ProvisioningActionHeal:
		return "Heal"
	default:
		return "Unknown"
	}
}

// EvaluateProvisioningAction returns the action to take based on the ElastiCache state and elapsed time.
// Order is significant: DELETING is always Wait; CREATE_FAILED within the deadline is Heal; the deadline
// Timeout applies ONLY to in-progress provisioning states (CREATING/CREATE_FAILED) — an already-provisioned
// instance that is modifying/snapshotting/available is never failed by it; everything else is Wait.
func EvaluateProvisioningAction(cacheState string, createdAt, now time.Time) ProvisioningAction {
	age := now.Sub(createdAt)

	if cacheState == ElastiCache_DELETING {
		return ProvisioningActionWait
	}
	if cacheState == ElastiCache_CREATE_FAILED && age < ElastiCacheProvisioningTimeout {
		return ProvisioningActionHeal
	}
	if (cacheState == ElastiCache_CREATING || cacheState == ElastiCache_CREATE_FAILED) && age >= ElastiCacheProvisioningTimeout {
		return ProvisioningActionTimeout
	}
	return ProvisioningActionWait
}
