package meta

import (
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	builder := &ErrorHandlerBuilder{
		err: NewAzureNotFoundError(),
	}
	builder.calculate()

	assert.True(t, builder.notFound)
	assert.False(t, builder.tooManyRequests)
	assert.Empty(t, builder.message)
	assert.Empty(t, builder.condition.Message)
	assert.Empty(t, builder.condition.Reason)
	assert.Nil(t, builder.successError)
	assert.Empty(t, builder.statusState)
}

func TestIsTooManyRequests(t *testing.T) {
	builder := &ErrorHandlerBuilder{
		err: NewAzureTooManyRequestsError(),
	}
	builder.calculate()

	assert.False(t, builder.notFound)
	assert.True(t, builder.tooManyRequests)
	assert.Empty(t, builder.message)
	assert.Empty(t, builder.condition.Message)
	assert.Empty(t, builder.condition.Reason)
	assert.Nil(t, builder.successError)
	assert.Empty(t, builder.statusState)
}

func TestAuthorizationFailed(t *testing.T) {
	builder := &ErrorHandlerBuilder{
		err: NewAzureAuthorizationFailedError(),
	}
	builder.calculate()

	assert.False(t, builder.notFound)
	assert.False(t, builder.tooManyRequests)
	assert.Equal(t, AuthorizationFailedMessage, builder.message)
	assert.Equal(t, AuthorizationFailedMessage, builder.condition.Message)
	assert.Equal(t, cloudcontrolv1beta1.ReasonUnauthorized, builder.condition.Reason)
	assert.Equal(t, builder.unauthorizedError, builder.successError)
	assert.Equal(t, cloudcontrolv1beta1.StateWarning, builder.statusState)
}

func TestAuthenticationFailed(t *testing.T) {
	builder := &ErrorHandlerBuilder{
		err: NewAzureAuthenticationFailedError(),
	}
	builder.calculate()

	assert.False(t, builder.notFound)
	assert.False(t, builder.tooManyRequests)
	assert.Equal(t, MissingServicePrincipalMessage, builder.message)
	assert.Equal(t, MissingServicePrincipalMessage, builder.condition.Message)
	assert.Equal(t, cloudcontrolv1beta1.ReasonUnauthenticated, builder.condition.Reason)
	assert.Equal(t, builder.unauthenticatedError, builder.successError)
	assert.Equal(t, cloudcontrolv1beta1.StateWarning, builder.statusState)
}

func TestConflict(t *testing.T) {
	builder := &ErrorHandlerBuilder{
		err: NewAzureConflictError(),
	}
	builder.calculate()

	assert.False(t, builder.notFound)
	assert.False(t, builder.tooManyRequests)
	assert.Equal(t, ConflictMessage, builder.message)
	assert.Equal(t, ConflictMessage, builder.condition.Message)
	assert.Equal(t, cloudcontrolv1beta1.ReasonConflict, builder.condition.Reason)
	assert.Equal(t, builder.conflictError, builder.successError)
	assert.Equal(t, cloudcontrolv1beta1.StateWarning, builder.statusState)
}
func TestConflictMessageOverride(t *testing.T) {
	builder := &ErrorHandlerBuilder{
		err: NewAzureConflictError(),
	}
	builder.WithConflictMessage("ConflictOverride").calculate()

	assert.False(t, builder.notFound)
	assert.False(t, builder.tooManyRequests)
	assert.Equal(t, "ConflictOverride", builder.message)
	assert.Equal(t, "ConflictOverride", builder.condition.Message)
	assert.Equal(t, cloudcontrolv1beta1.ReasonConflict, builder.condition.Reason)
	assert.Equal(t, builder.conflictError, builder.successError)
	assert.Equal(t, cloudcontrolv1beta1.StateWarning, builder.statusState)
}

func TestDefaults(t *testing.T) {
	builder := &ErrorHandlerBuilder{
		err: errors.New("some generic error"),
	}
	builder.WithDefaultReason("DefaultReason").
		WithDefaultMessage("DefaultMessage").
		calculate()

	assert.False(t, builder.notFound)
	assert.False(t, builder.tooManyRequests)
	assert.Equal(t, "DefaultMessage", builder.message)
	assert.Equal(t, "DefaultMessage", builder.condition.Message)
	assert.Equal(t, "DefaultReason", builder.condition.Reason)
	assert.Equal(t, composed.StopAndForget, builder.successError)
	assert.Equal(t, cloudcontrolv1beta1.StateError, builder.statusState)
}
