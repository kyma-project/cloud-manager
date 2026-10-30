package v2

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awserrorhandling "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/errorhandling"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func vpcLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.ObjAsIpRange().Status.VpcId) == 0 {
		return nil, nil
	}

	vpcId := state.ObjAsIpRange().Status.VpcId
	vpc, err := state.awsClient.DescribeVpc(ctx, vpcId)

	// The VPC is reported as missing either as an empty result (vpc == nil, no error)
	// or as an AWS InvalidVpcID.NotFound error.
	if vpc == nil || awsmeta.IsNotFound(err) {
		// On deletion a missing VPC is the terminal success condition: every AWS
		// resource that lived in it (subnets, CIDR association) is already gone, so
		// pass through with state.vpc == nil and let the delete actions and finalizer
		// removal run.
		if composed.IsMarkedForDeletion(state.Obj()) {
			logger.Info("VPC already gone, proceeding with KCP IpRange deletion")
			return nil, ctx
		}

		// On create/update the VPC must exist. Back off with a real delay - NEVER an
		// immediate requeue (StopWithRequeue is RequeueAfter 1ns and hot-loops the
		// reconcile workers), and set the Error condition.
		logger.Error(fmt.Errorf("vpc %s not found", vpcId), "VPC not found for KCP IpRange")
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
				Message: fmt.Sprintf("AWS VPC %s not found", vpcId),
			}).
			ErrorLogMessage("Error patching KCP IpRange status when VPC not found on load").
			SuccessLogMsg("Delaying KCP IpRange reconciliation with AWS VPC not found").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, st)
	}

	// A genuine transient/other AWS error keeps the existing retryable handling.
	if err != nil {
		return awserrorhandling.HandleError(ctx, err, state, "KCP IpRange on VPC load",
			cloudcontrolv1beta1.ReasonUnknown, "Error loading AWS VPC"), ctx
	}

	state.vpc = vpc

	logger = logger.WithValues("vpcId", ptr.Deref(state.vpc.VpcId, ""))
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
