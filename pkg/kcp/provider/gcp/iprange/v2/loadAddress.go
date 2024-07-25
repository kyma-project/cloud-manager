package v2

import (
	"context"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"google.golang.org/api/googleapi"
)

func loadAddress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Loading GCP Address")

	//Get from GCP.
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork
	name := ipRange.Spec.RemoteRef.Name
	if ipRange.Status.Id != "" {
		name = ipRange.Status.Id
	}

	addr, err := state.computeClient.GetIpRange(ctx, project, name)
	if err != nil {

		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			if e.Code == 404 {
				state.address = nil
				return nil, nil
			}
		}

		return composed.UpdateStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "Error getting Addresses from GCP",
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Error getting Addresses from GCP").
			Run(ctx, state)
	}

	//Check whether the IPRange is in the same VPC as that of the SKR.
	if !strings.HasSuffix(addr.Network, vpc) {
		return composed.UpdateStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "IPRange with the same name exists in another VPC.",
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("GCP - IPRange name conflict").
			Run(ctx, state)
	}
	state.address = addr

	return nil, nil
}
