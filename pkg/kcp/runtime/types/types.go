package types

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
)

type State interface {
	composed.State
	ObjAsRuntime() *infrastructuremanagerv1.Runtime
	Subscription() *cloudcontrolv1beta1.Subscription
	VpcNetwork() *cloudcontrolv1beta1.VpcNetwork

	SecurityServiceEnabledOnSubscription() bool
	SecurityServiceEnabledOnSubscriptionPredicate(ctx context.Context, st composed.State) bool

	SecurityDataSourceEnabledOnRuntime() bool
	SecurityDataSourceEnabledOnRuntimePredicate(ctx context.Context, st composed.State) bool

	PatchStatusAnnotations(ctx context.Context, newStatus, newMessage string, observedGeneration int64) (error, context.Context)

	SecurityDesiredState() SecurityDesiredState
}

type SecurityDesiredState interface {
	RuntimeId() string
	SubscriptionId() string
	EnabledOnRuntime() bool
	EnabledOnSubscription() bool
}
