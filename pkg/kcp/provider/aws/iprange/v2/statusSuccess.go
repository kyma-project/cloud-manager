package v2

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func statusSuccess(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.ReadyState
	state.ObjAsIpRange().Status.Subnets = pie.Map(state.cloudResourceSubnets, func(s ec2Types.Subnet) cloudcontrolv1beta1.IpRangeSubnet {
		return cloudcontrolv1beta1.IpRangeSubnet{
			Id:    ptr.Deref(s.SubnetId, ""),
			Zone:  ptr.Deref(s.AvailabilityZone, ""),
			Range: ptr.Deref(s.CidrBlock, ""),
		}
	})
	return composed.PatchStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "Additional IpRange(s) are provisioned",
		}).
		ErrorLogMessage("Error patching KCP IpRange status with ready state").
		SuccessLogMsg("Forgetting KCP IpRange with ready state").
		Run(ctx, state)
}
