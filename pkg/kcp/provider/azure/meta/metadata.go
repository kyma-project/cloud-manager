package meta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"io"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"regexp"
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
	VnetAddressSpacesOverlap                             = "VnetAddressSpacesOverlap"
	VnetAddressSpacesOverlapMessage                      = "Cannot create or update peering. Virtual networks cannot be peered because their address spaces overlap. "
	InvalidAuthenticationTokenTenant                     = "InvalidAuthenticationTokenTenant"
	InvalidAuthenticationTokenTenantMessage              = "Authentication failed"
	MissingServicePrincipalMessage                       = "The client application is missing service principal in the remote tenant"
	Conflict                                             = "Conflict"
	ConflictMessage                                      = "Another Azure operation is pending for requested object"
)

func IsTooManyRequests(err error) bool {
	var respErr *azcore.ResponseError

	// https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/request-limits-and-throttling
	return errors.As(err, &respErr) && respErr.StatusCode == http.StatusTooManyRequests
}

func IsConflictError(err error) bool {
	var respErr *azcore.ResponseError
	return errors.As(err, &respErr) && respErr.StatusCode == http.StatusConflict
}

func NewAzureNotFoundError() error {
	return &azcore.ResponseError{
		ErrorCode:  "ResourceNotFound",
		StatusCode: http.StatusNotFound,
	}
}
func NewAzureTooManyRequestsError() error {
	return &azcore.ResponseError{
		StatusCode: http.StatusTooManyRequests,
	}
}
func NewAzureAuthorizationFailedError() error {
	return &azcore.ResponseError{
		ErrorCode:  AuthorizationFailed,
		StatusCode: http.StatusUnauthorized,
	}
}

func NewAzureAuthenticationFailedError() error {
	return &azidentity.AuthenticationFailedError{}
}

func NewAzureConflictError() error {
	return &azcore.ResponseError{
		ErrorCode:  Conflict,
		StatusCode: http.StatusConflict,
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

func IsUnauthorized(err error) bool {
	var respErr *azcore.ResponseError

	if ok := errors.As(err, &respErr); ok {
		return respErr.ErrorCode == "AuthorizationFailed"
	}

	return false
}

func IsUnauthenticated(err error) bool {
	var auth *azidentity.AuthenticationFailedError
	if ok := errors.As(err, &auth); ok {
		return true
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

func GetErrorMessage(err error, def string) (string, bool) {
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
		case VnetAddressSpacesOverlap:
			r := regexp.MustCompile(`Overlapping address prefixes:[^\"]*`)
			return VnetAddressSpacesOverlapMessage + r.FindString(respErr.Error()), true
		case InvalidAuthenticationTokenTenant:
			return InvalidAuthenticationTokenTenantMessage, true
		case Conflict:
			return ConflictMessage, true
		default:
			msg := extractAzureErrorMessage(respErr)
			if msg != "" {
				return msg, true
			}
		}
	}

	if IsUnauthenticated(err) {
		return MissingServicePrincipalMessage, true
	}

	return def, false
}

type ErrorHandlerBuilder struct {
	err                    error
	obj                    composed.ObjWithConditionsAndState
	condition              metav1.Condition
	statusState            cloudcontrolv1beta1.StatusState
	message                string
	successError           error
	defaultReason          string
	defaultMessage         string
	tooManyRequests        bool
	notFound               bool
	tooManyRequestsMessage string
	updateStatusMessage    string
	notFoundMessage        string
	conflictMessage        string
	unauthorizedError      error
	unauthenticatedError   error
	conflictError          error
}

func HandleError(err error, obj composed.ObjWithConditionsAndState) *ErrorHandlerBuilder {
	return &ErrorHandlerBuilder{
		err: err,
		obj: obj,
	}
}

func (b *ErrorHandlerBuilder) WithDefaultReason(reason string) *ErrorHandlerBuilder {
	b.defaultReason = reason
	return b
}

func (b *ErrorHandlerBuilder) WithDefaultMessage(message string) *ErrorHandlerBuilder {
	b.defaultMessage = message
	return b
}

func (b *ErrorHandlerBuilder) WithTooManyRequestsMessage(message string) *ErrorHandlerBuilder {
	b.tooManyRequestsMessage = message
	return b
}

func (b *ErrorHandlerBuilder) WithConflictMessage(message string) *ErrorHandlerBuilder {
	b.conflictMessage = message
	return b
}

func (b *ErrorHandlerBuilder) WithUpdateStatusMessage(message string) *ErrorHandlerBuilder {
	b.updateStatusMessage = message
	return b
}

func (b *ErrorHandlerBuilder) WithNotFoundMessage(message string) *ErrorHandlerBuilder {
	b.notFoundMessage = message
	return b
}

func (b *ErrorHandlerBuilder) setDefaults() {
	b.unauthenticatedError = composed.StopWithRequeueDelay(util.Timing.T300000ms())
	b.unauthorizedError = composed.StopWithRequeueDelay(util.Timing.T300000ms())
	b.conflictError = composed.StopWithRequeueDelay(util.Timing.T60000ms())

	if b.conflictMessage == "" {
		b.conflictMessage = ConflictMessage
	}
}

func (b *ErrorHandlerBuilder) Run(ctx context.Context, state composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)

	b.calculate()

	if b.notFound {
		logger.Info(b.notFoundMessage)
		return nil, ctx
	}

	if b.tooManyRequests {
		return composed.LogErrorAndReturn(b.err,
			b.tooManyRequestsMessage,
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}

	if b.err != nil {
		logger.Error(b.err, b.message)
	}

	changed := meta.SetStatusCondition(b.obj.Conditions(), b.condition)

	if b.obj.State() != string(b.statusState) {
		b.obj.SetState(string(b.statusState))
		changed = true
	}

	if changed {
		return composed.UpdateStatus(b.obj).
			ErrorLogMessage(b.updateStatusMessage).
			SuccessLogMsg(fmt.Sprintf("Status successfully updated with status '%s' and message '%s'", b.statusState, b.message)).
			SuccessError(b.successError).
			Run(ctx, state)
	}

	return b.successError, ctx
}

func (b *ErrorHandlerBuilder) calculate() {
	b.setDefaults()

	if IsNotFound(b.err) {
		b.notFound = true
		return
	}

	if IsTooManyRequests(b.err) {
		b.tooManyRequests = true
		return
	}

	var isWarning bool
	b.message, isWarning = GetErrorMessage(b.err, b.defaultMessage)

	b.condition = metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeError,
		Status:  metav1.ConditionTrue,
		Reason:  b.defaultReason,
		Message: b.message,
	}

	b.successError = composed.StopAndForget
	if IsUnauthorized(b.err) {
		b.condition.Reason = cloudcontrolv1beta1.ReasonUnauthorized
		b.successError = b.unauthorizedError
	}

	if IsUnauthenticated(b.err) {
		b.condition.Reason = cloudcontrolv1beta1.ReasonUnauthenticated
		b.successError = b.unauthenticatedError
	}

	if IsConflictError(b.err) {
		b.condition.Message = b.conflictMessage
		b.message = b.conflictMessage
		b.condition.Reason = cloudcontrolv1beta1.ReasonConflict
		b.successError = b.conflictError
	}

	if isWarning {
		b.statusState = cloudcontrolv1beta1.StateWarning
	} else {
		b.statusState = cloudcontrolv1beta1.StateError
	}
}

type azureErrorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func extractAzureErrorMessage(respErr *azcore.ResponseError) string {

	if respErr == nil || respErr.RawResponse.Body == nil {
		return ""
	}

	defer respErr.RawResponse.Body.Close()

	body, _ := io.ReadAll(respErr.RawResponse.Body)

	var azErr azureErrorBody
	if json.Unmarshal(body, &azErr) != nil {
		return ""
	}

	return azErr.Error.Message
}
