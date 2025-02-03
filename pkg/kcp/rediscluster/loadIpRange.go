package rediscluster

import (
	"context"
	"fmt"

	types2 "github.com/kyma-project/cloud-manager/pkg/kcp/rediscluster/types"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(types2.State)
	logger := composed.LoggerFromCtx(ctx)

	redisCluster := state.ObjAsRedisCluster()
	ipRangeName := redisCluster.Spec.IpRange.Name

	ipRange := &cloudcontrolv1beta1.IpRange{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      ipRangeName,
	}, ipRange)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading referred IpRange", composed.StopWithRequeue, ctx)
	}

	if apierrors.IsNotFound(err) {
		logger.
			WithValues("ipRange", ipRangeName).
			Error(err, "Referred IpRange does not exist")
		meta.SetStatusCondition(redisCluster.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonInvalidIpRangeReference,
			Message: fmt.Sprintf("Referred IpRange %s/%s does not exist", state.Obj().GetNamespace(), ipRangeName),
		})
		redisCluster.SetStatusStateToError()
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating RedisCluster status after referred IpRange not found", composed.StopWithRequeue, ctx)
		}

		return composed.StopAndForget, nil
	}

	state.SetIpRange(ipRange)

	return nil, nil
}
