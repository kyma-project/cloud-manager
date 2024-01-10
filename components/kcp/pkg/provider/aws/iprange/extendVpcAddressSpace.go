package iprange

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-resources-manager/components/lib/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func extendVpcAddressSpace(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	if state.associatedCidrBlock != nil {
		return nil, nil
	}

	logger.Info("Associating vpc cidr block")

	_, err := state.client.AssociateVpcCidrBlock(ctx, pointer.StringDeref(state.vpc.VpcId, ""), state.ObjAsIpRange().Spec.Cidr)

	if err != nil {
		logger.Error(err, "Error associating vpc cidr block")
		meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonFailedExtendingVpcAddressSpace,
			Message: "Failed extending vpc address space",
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating status due to failed extending vpc address space", composed.StopWithRequeue, nil)
		}

		return composed.StopAndForget, nil
	}

	return nil, nil
}
