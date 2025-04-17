package gcprediscluster

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
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

	gcpSubnet := &cloudresourcesv1beta1.GcpSubnet{}
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
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonSubnetNotFound,
			Message: fmt.Sprintf("Referred GcpSubnet %s/%s does not exist", state.Obj().GetNamespace(), subnetName),
		})
		redisCluster.Status.State = cloudresourcesv1beta1.StateError
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating GcpRedisCluster status after referred GcpSubnet not found", composed.StopWithRequeue, ctx)
		}

		return composed.StopAndForget, nil
	}

	state.SkrSubnet = gcpSubnet

	return nil, nil
}
