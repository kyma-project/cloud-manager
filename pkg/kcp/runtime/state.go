package runtime

import (
	"context"
	"fmt"
	"strconv"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	runtimetypes "github.com/kyma-project/cloud-manager/pkg/kcp/runtime/types"
)

var _ runtimetypes.State = &State{}

type State struct {
	composed.State

	subscription *cloudcontrolv1beta1.Subscription
	vpcNetwork   *cloudcontrolv1beta1.VpcNetwork

	// securityServiceEnabledOnSubscription true if any runtime in the subscription has security enabled
	// indicates that security scanning services producing findings has to
	securityServiceEnabledOnSubscription bool

	// securityDataSourceEnabledOnRuntime trye if current runtime resource has enabled security
	// indicated that runtime network related stuff (flow logs) should be enabled
	securityDataSourceEnabledOnRuntime bool
}

func newState(baseState composed.State) *State {
	return &State{
		State: baseState,
	}
}

func (s *State) ObjAsRuntime() *infrastructuremanagerv1.Runtime {
	return s.Obj().(*infrastructuremanagerv1.Runtime)
}

func (s *State) Subscription() *cloudcontrolv1beta1.Subscription {
	return s.subscription
}

func (s *State) VpcNetwork() *cloudcontrolv1beta1.VpcNetwork {
	return s.vpcNetwork
}

func (s *State) SecurityServiceEnabledOnSubscription() bool {
	return s.securityServiceEnabledOnSubscription
}

func (s *State) SecurityDataSourceEnabledOnRuntime() bool {
	return s.securityDataSourceEnabledOnRuntime
}

func (s *State) SecurityServiceEnabledOnSubscriptionPredicate(_ context.Context, st composed.State) bool {
	return s.SecurityServiceEnabledOnSubscription()
}

func (s *State) SecurityDataSourceEnabledOnRuntimePredicate(_ context.Context, st composed.State) bool {
	return s.SecurityDataSourceEnabledOnRuntime()
}

func (s *State) PatchStatusAnnotations(ctx context.Context, newStatus, newMessage string, observedGeneration int64) (error, context.Context) {
	_, err := composed.PatchObjMergeAnnotations(ctx, s.ObjAsRuntime(), s.Cluster().K8sClient(), map[string]string{
		cloudcontrolv1beta1.RuntimeSecurityStatusAnnotation:             newStatus,
		cloudcontrolv1beta1.RuntimeSecurityMessageAnnotation:            newMessage,
		cloudcontrolv1beta1.RuntimeSecurityObservedGenerationAnnotation: strconv.FormatInt(observedGeneration, 10),
		cloudcontrolv1beta1.RuntimeSecurityLastReconcileTime:            time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, fmt.Sprintf("Failed to patch Runtime status annotations to %s", newStatus), composed.StopWithRequeueDelay(rate.Slow1s.When(s.Obj())), ctx)
	}

	return nil, ctx
}

func (s *State) SecurityDesiredState() runtimetypes.SecurityDesiredState {
	if !composed.IsObjLoaded(context.TODO(), s) || s.subscription == nil {
		return nil
	}
	return &securityDesiredState{
		runtimeId:             s.ObjAsRuntime().Name,
		subscriptionId:        s.subscription.Name,
		enabledOnRuntime:      s.securityDataSourceEnabledOnRuntime,
		enabledOnSubscription: s.securityServiceEnabledOnSubscription,
	}
}

// ============================================

type securityDesiredState struct {
	runtimeId             string
	subscriptionId        string
	enabledOnRuntime      bool
	enabledOnSubscription bool
}

var _ runtimetypes.SecurityDesiredState = (*securityDesiredState)(nil)

func (s securityDesiredState) RuntimeId() string {
	return s.runtimeId
}

func (s securityDesiredState) SubscriptionId() string {
	return s.subscriptionId
}

func (s securityDesiredState) EnabledOnRuntime() bool {
	return s.enabledOnRuntime
}

func (s securityDesiredState) EnabledOnSubscription() bool {
	return s.enabledOnSubscription
}
