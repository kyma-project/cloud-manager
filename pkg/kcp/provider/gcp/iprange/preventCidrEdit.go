package iprange

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// preventCidrEdit prevents CIDR changes after the IpRange is in Ready state.
// Once an IpRange is Ready and allocated in GCP, changing the CIDR would require
// recreating the resource, which could break existing dependencies.
func preventCidrEdit(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	ipRange := state.ObjAsIpRange()

	// Skip if spec.cidr is not set
	if len(ipRange.Spec.Cidr) == 0 {
		return nil, nil
	}

	// Skip if status.cidr is not set (not yet allocated)
	if len(ipRange.Status.Cidr) == 0 {
		return nil, nil
	}

	// Check if CIDR was changed
	if ipRange.Spec.Cidr != ipRange.Status.Cidr {
		// Only prevent changes if resource is Ready
		readyCondition := meta.FindStatusCondition(*ipRange.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
		if readyCondition != nil && readyCondition.Status == metav1.ConditionTrue {
			return composed.PatchStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCidrCanNotChange,
					Message: "CIDR cannot be changed after IpRange is Ready",
				}).
				ErrorLogMessage("Error patching IpRange status after CIDR change in Ready state").
				SuccessError(composed.StopAndForget).
				SuccessLogMsg("Rejecting CIDR change for Ready IpRange").
				Run(ctx, st)
		}
	}

	return nil, nil
}
