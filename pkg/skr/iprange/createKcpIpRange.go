package iprange

import (
	"context"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			Name:      uuid.NewString(),
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      state.Name().Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: state.Name().Namespace,
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

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpIpRange)
	if err != nil {
		logger.Error(err, "Error creating KCP IpRange")
		return composed.StopWithRequeue, nil
	}
	logger.
		WithValues("kcpIpRangeName", state.KcpIpRange.Name).
		Info("KCP IpRange created")

	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeSubmitted,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonSubmissionSucceeded,
			Message: "Resource is submitted for provisioning",
		}).
		ErrorLogMessage("Error updating IpRange status with submitted condition").
		Run(ctx, state)
}
