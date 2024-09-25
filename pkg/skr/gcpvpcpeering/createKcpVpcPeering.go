package gcpvpcpeering

import (
	"context"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.KcpVpcPeering != nil {
		return nil, nil
	}

	obj := state.ObjAsGcpVpcPeering()

	// If there is no guid, generate one before creating the KCP VpcPeering
	if obj.Status.Id == "" {
		statusId := uuid.NewString()

		obj.Status.Id = statusId
		err := state.UpdateObjStatus(ctx)

		// If there is too many requests to the API, log and requeue the request
		if err != nil {
			return composed.LogErrorAndReturn(err, "[SKR GcpVpcPeering] Error updating status with ID "+err.Error(), composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
	}

	state.KcpVpcPeering = &cloudcontrolv1beta1.VpcPeering{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      obj.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: obj.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.VpcPeeringSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: obj.Namespace,
				Name:      obj.Name,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			VpcPeering: &cloudcontrolv1beta1.VpcPeeringInfo{
				Gcp: &cloudcontrolv1beta1.GcpVpcPeering{
					ImportCustomRoutes: obj.Spec.ImportCustomRoutes,
					RemoteVpc:          obj.Spec.RemoteVpc,
					RemoteProject:      obj.Spec.RemoteProject,
					RemotePeeringName:  obj.Spec.RemotePeeringName,
				},
			},
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpVpcPeering)

	if err != nil {
		return composed.LogErrorAndReturn(err, "[SKR GcpVpcPeering] Error creating KCP VpcPeering", composed.StopWithRequeue, ctx)
	}

	logger.Info("[SKR GcpVpcPeering] Created KCP VpcPeering")

	return nil, nil
}
