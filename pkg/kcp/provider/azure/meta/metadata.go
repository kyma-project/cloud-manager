package meta

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

const (
	RemotePeeringIsDisconnected                      = "RemotePeeringIsDisconnected"
	RemotePeeringIsDisconnectedMessage               = "Cannot create or update peering because remote peering referencing parent virtual network is in Disconnected state. Update or re-create the remote peering to get it back to Initiated state. Peering gets Disconnected when remote vnet or remote peering is deleted and re-created"
	AnotherPeeringAlreadyReferencesRemoteVnet        = "AnotherPeeringAlreadyReferencesRemoteVnet"
	AnotherPeeringAlreadyReferencesRemoteVnetMessage = "Peering already references remote virtual network. Cannot add another peering referencing the same remote virtual network."
	AuthorizationFailed                              = "AuthorizationFailed"
	AuthorizationFailedMessage                       = "Not authorized to perform action."
)

func TooManyRequests(err error) bool {
	var respErr *azcore.ResponseError

	// https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/request-limits-and-throttling
	return errors.As(err, &respErr) && respErr.StatusCode == 429
}

func IsNotFound(err error) bool {
	var respErr *azcore.ResponseError

	if ok := errors.As(err, &respErr); ok {
		if respErr.StatusCode == 404 {
			return true
		}
		return respErr.ErrorCode == "ResourceNotFound"
	}

	return false
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

func GetErrorMessage(err error) string {
	var respErr *azcore.ResponseError

	if errors.As(err, &respErr) {
		switch respErr.ErrorCode {
		case RemotePeeringIsDisconnected:
			return RemotePeeringIsDisconnectedMessage
		case AnotherPeeringAlreadyReferencesRemoteVnet:
			return AnotherPeeringAlreadyReferencesRemoteVnetMessage
		case AuthorizationFailed:
			return AuthorizationFailedMessage
		}
	}

	return fmt.Sprintf("Failed creating VpcPeerings %s", err)
}
