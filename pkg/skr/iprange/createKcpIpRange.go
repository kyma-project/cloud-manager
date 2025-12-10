package iprange

import (
	"context"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func createKcpIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR IpRange is marked for deletion, do not create mirror in KCP
		return nil, nil
	}

	if state.KcpIpRange != nil {
		// mirror IpRange in KCP is already created
		return nil, nil
	}

	logger.Info("Creating KCP IpRange")

	state.KcpIpRange = &cloudcontrolv1beta1.IpRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      state.ObjAsIpRange().Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:   state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName: state.Name().Name,
				common.LabelKymaModule:              common.FieldOwner,
			},
		},
		Spec: cloudcontrolv1beta1.IpRangeSpec{
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: state.ObjAsIpRange().Namespace,
				Name:      state.ObjAsIpRange().Name,
			},
			Cidr: state.ObjAsIpRange().Spec.Cidr,
		},
	}
	if state.Provider != nil && *state.Provider == cloudcontrolv1beta1.ProviderAzure {
		state.KcpIpRange.Spec.Network = &klog.ObjectRef{
			Name: common.KcpNetworkCMCommonName(state.KymaRef.Name),
		}
	}
	if state.Provider != nil && *state.Provider != cloudcontrolv1beta1.ProviderAzure {
		state.KcpIpRange.Spec.Network = &klog.ObjectRef{
			Name: common.KcpNetworkKymaCommonName(state.KymaRef.Name),
		}
	}
	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpIpRange)
	if err != nil {
		logger.Error(err, "Error creating KCP IpRange")
		return composed.StopWithRequeue, nil
	}
	logger.
		WithValues("kcpIpRangeName", state.KcpIpRange.Name).
		Info("KCP IpRange created")

	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeSubmitted,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonSubmissionSucceeded,
			Message: "Resource is submitted for provisioning",
		}).
		DeriveStateFromConditions(state.MapConditionToState()).
		ErrorLogMessage("Error updating IpRange status with submitted condition").
		SuccessError(composed.StopWithRequeueDelay(100*time.Millisecond)).
		Run(ctx, state)
}
