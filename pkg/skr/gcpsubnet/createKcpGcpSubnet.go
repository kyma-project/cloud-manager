package gcpsubnet

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func createKcpGcpSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpGcpSubnet != nil {
		return nil, ctx
	}

	gcpSubnet := state.ObjAsGcpSubnet()

	state.KcpGcpSubnet = &cloudcontrolv1beta1.GcpSubnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gcpSubnet.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				common.LabelKymaModule: "cloud-manager",
			},
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:   state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName: gcpSubnet.Name,
			},
		},
		Spec: cloudcontrolv1beta1.GcpSubnetSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: gcpSubnet.Namespace,
				Name:      gcpSubnet.Name,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			Network: &klog.ObjectRef{
				Name: common.KcpNetworkKymaCommonName(state.KymaRef.Name),
			},
			Cidr:    gcpSubnet.Spec.Cidr,
			Purpose: cloudcontrolv1beta1.GcpSubnetPurpose_PRIVATE,
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpGcpSubnet)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP GcpSubnet", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP GcpSubnet")

	gcpSubnet.Status.State = cloudresourcesv1beta1.StateCreating
	return composed.UpdateStatus(gcpSubnet).
		ErrorLogMessage("Error setting Creating state on GcpSubnet").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
