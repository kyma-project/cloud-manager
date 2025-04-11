package rediscluster

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	redisCluster := state.ObjAsGcpRedisCluster()
	subnetName := redisCluster.Spec.Subnet.Name

	gcpSubnet := &cloudcontrolv1beta1.GcpSubnet{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      subnetName,
	}, gcpSubnet)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading referred GcpSubnet", composed.StopWithRequeue, ctx)
	}

	if apierrors.IsNotFound(err) {
		logger.
			WithValues("subnet", subnetName).
			Error(err, "Referred GcpSubnet does not exist")
		meta.SetStatusCondition(redisCluster.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonNotFound,
			Message: fmt.Sprintf("Referred GcpSubnet %s/%s does not exist", state.Obj().GetNamespace(), subnetName),
		})
		redisCluster.SetStatusStateToError()
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating GcpRedisCluster status after referred GcpSubnet not found", composed.StopWithRequeue, ctx)
		}

		return composed.StopAndForget, nil
	}

	state.SetSubnet(gcpSubnet)

	return nil, nil
}
