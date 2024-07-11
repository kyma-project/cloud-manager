package v1

import (
	"context"
	"fmt"
	"k8s.io/utils/ptr"

	"github.com/3th1nk/cidr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkCidrOverlap(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	rangeCidr, err := cidr.Parse(state.ObjAsIpRange().Spec.Cidr)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error parsing IpRange .spec.cidr", composed.StopWithRequeue, ctx)
	}

	for _, set := range state.vpc.CidrBlockAssociationSet {
		cdr, err := cidr.Parse(ptr.Deref(set.CidrBlock, ""))
		if err != nil {
			logger.Error(err, "Error parsing AWS CIDR: %w", err)
			continue
		}

		if util.CidrEquals(rangeCidr.CIDR(), cdr.CIDR()) {
			continue
		}

		if util.CidrOverlap(rangeCidr.CIDR(), cdr.CIDR()) {
			meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonCidrOverlap,
				Message: fmt.Sprintf("CIDR overlaps with VPC adress range cidr %s", ptr.Deref(set.CidrBlock, "")),
			})
			err := state.UpdateObjStatus(ctx)
			if err != nil {
				return composed.LogErrorAndReturn(err, "Error updating IpRange status due to cidr overlap", composed.StopWithRequeue, ctx)
			}

			return composed.StopAndForget, nil
		}
	}

	return nil, nil
}
