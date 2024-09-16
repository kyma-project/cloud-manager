package meta

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"net/http"
)

const (
	RemotePeeringIsDisconnected                          = "RemotePeeringIsDisconnected"
	RemotePeeringIsDisconnectedMessage                   = "Cannot create or update peering because remote peering referencing parent virtual network is in Disconnected state. Update or re-create the remote peering to get it back to Initiated state. Peering gets Disconnected when remote vnet or remote peering is deleted and re-created"
	AnotherPeeringAlreadyReferencesRemoteVnet            = "AnotherPeeringAlreadyReferencesRemoteVnet"
	AnotherPeeringAlreadyReferencesRemoteVnetMessage     = "Peering already references remote virtual network. Cannot add another peering referencing the same remote virtual network."
	AuthorizationFailed                                  = "AuthorizationFailed"
	AuthorizationFailedMessage                           = "Not authorized to perform action."
	VnetAddressSpaceOverlapsWithAlreadyPeeredVnet        = "VnetAddressSpaceOverlapsWithAlreadyPeeredVnet"
	VnetAddressSpaceOverlapsWithAlreadyPeeredVnetMessage = "Cannot create or update peering. Virtual networks cannot be peered because address space of the first virtual network overlaps with address space of third virtual network already peered with the second virtual network."
	InvalidResourceName                                  = "InvalidResourceName"
	InvalidResourceNameMessage                           = "Resource name is invalid. The name can be up to 80 characters long. It must begin with a word character, and it must end with a word character. The name may contain word characters or '.', '-'."
)

func IsTooManyRequests(err error) bool {
	var respErr *azcore.ResponseError

	// https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/request-limits-and-throttling
	return errors.As(err, &respErr) && respErr.StatusCode == http.StatusTooManyRequests
}

func NewAzureNotFoundError() error {
	return &azcore.ResponseError{
		ErrorCode:  "ResourceNotFound",
		StatusCode: http.StatusNotFound,
	}
}

func IgnoreNotFoundError(err error) error {
	if IsNotFound(err) {
		return nil
	}
	return err
}

func IsNotFound(err error) bool {
	var respErr *azcore.ResponseError

	if ok := errors.As(err, &respErr); ok {
		if respErr.StatusCode == http.StatusNotFound {
			return true
		}
		return respErr.ErrorCode == "ResourceNotFound"
	}

	return false
}

func ErrorToRequeueResponse(err error) error {
	if IsTooManyRequests(err) {
		return composed.StopWithRequeueDelay(util.Timing.T60000ms())
	}
	return composed.StopWithRequeueDelay(util.Timing.T300000ms())
}

func LogErrorAndReturn(err error, msg string, ctx context.Context) (error, context.Context) {
	result := ErrorToRequeueResponse(err)
	return composed.LogErrorAndReturn(err, msg, result, ctx)
}

func GetErrorMessage(err error) (string, bool) {
	var respErr *azcore.ResponseError

	if errors.As(err, &respErr) {
		switch respErr.ErrorCode {
		case RemotePeeringIsDisconnected:
			return RemotePeeringIsDisconnectedMessage, true
		case AnotherPeeringAlreadyReferencesRemoteVnet:
			return AnotherPeeringAlreadyReferencesRemoteVnetMessage, true
		case AuthorizationFailed:
			return AuthorizationFailedMessage, true
		case VnetAddressSpaceOverlapsWithAlreadyPeeredVnet:
			return VnetAddressSpaceOverlapsWithAlreadyPeeredVnetMessage, true
		case InvalidResourceName:
			return InvalidResourceNameMessage, true
		}
	}

	return fmt.Sprintf("Failed creating VpcPeerings %s", err), false
}
