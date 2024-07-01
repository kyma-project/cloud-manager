package vpcpeering

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"time"
)

func checkNetworkTag(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if state.vpcPeering != nil {
		return nil, nil
	}

	var kv []any

	var shootNameTag *ec2Types.Tag
	for _, t := range state.remoteVpc.Tags {
		kv = append(kv, *t.Key, *t.Value)
		if pointer.StringDeref(t.Key, "") == "shoot-name" {
			shootNameTag = &t
		}
	}

	// If VpcNetwork is found but tags don't match user can recover by adding tag to remote VPC network so, we are
	// adding stop with requeue delay of one minute.
	if shootNameTag == nil || pointer.StringDeref(shootNameTag.Value, "") != state.Scope().Spec.ShootName {

		logger.Info("Loaded remote VPC Network have no matching tags", kv...)

		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcNetwork,
				Message: fmt.Sprintf("Loaded remote Vpc network has no matching tags"),
			}).
			ErrorLogMessage("Error updating VpcPeering status due to remote vpc network tag mismatch").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(time.Minute)).
			Run(ctx, state)
	}

	return nil, nil
}