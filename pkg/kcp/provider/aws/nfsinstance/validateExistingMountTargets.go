package nfsinstance

import (
	"context"
	"fmt"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// validateExistingMountTargets validates if there are any mount targets referring to a subnet not
// created by the IpRange. If any such mount target is found the NfsInstance status is set to error
// state and reconciliation stopped. All this is due to the fact that AWS EFS API allows only
// one mount target per availability zone. If we keep trying to create a mount target in the
// IpRange subnet (in some zone) and another mount target in same zone already exists we would get an error
// that can not be fixed by repeating the reconciliation loop. Finer kind of validation can be made
// by loading that subnet referred by the mount target and checking its zone, but following the idea
// that no one should mess with cloud resources this operator created and the fact there's a small number
// of zones and we want to create a mount target in each, most probably that foreign subnet is occupying
// a zone we have IpRange's subnet in. So to fail quickly, as soon as we detect a mount target with
// non IpRange subnet we will put the object in the failed state
func validateExistingMountTargets(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	var invalidMountTargets []efsTypes.MountTargetDescription

	for _, mt := range state.mountTargets {
		x := state.IpRange().Status.Subnets.SubnetById(ptr.Deref(mt.SubnetId, ""))
		if x == nil {
			invalidMountTargets = append(invalidMountTargets, mt)
		}
	}

	if len(invalidMountTargets) == 0 {
		return nil, nil
	}

	logger.WithValues(
		"invalidMountTargets",
		fmt.Sprintf("%v", pie.Map(invalidMountTargets, func(mt efsTypes.MountTargetDescription) string {
			return fmt.Sprintf(
				"(%s %s %s)",
				ptr.Deref(mt.MountTargetId, ""),
				ptr.Deref(mt.AvailabilityZoneName, ""),
				ptr.Deref(mt.SubnetId, ""),
			)
		})),
	).
		Info("Invalid mount targets")

	meta.SetStatusCondition(state.ObjAsNfsInstance().Conditions(), metav1.Condition{
		Type:   cloudcontrolv1beta1.ConditionTypeError,
		Status: metav1.ConditionTrue,
		Reason: cloudcontrolv1beta1.ReasonInvalidMountTargetsAlreadyExist,
		Message: fmt.Sprintf("Invalid mount targets already exist: %v", pie.Map(invalidMountTargets, func(mt efsTypes.MountTargetDescription) string {
			return ptr.Deref(mt.MountTargetId, "")
		})),
	})
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating NfsInstance status conditions after invalid mount targets found", composed.StopWithRequeue, ctx)
	}

	state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.ErrorState
	err = state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating NfsInstance status state after invalid mount targets found", composed.StopWithRequeue, ctx)
	}

	return composed.StopAndForget, nil
}
