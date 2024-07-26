package v2

import (
	"context"
	"errors"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
)

func loadAddress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()

	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpcName := gcpScope.VpcNetwork
	remoteName := GetIpRangeName(ipRange.GetName())
	remoteFallbackName := ipRange.Spec.RemoteRef.Name
	logger = logger.WithValues(
		"ipRange", ipRange.Name,
		"ipRangeRemoteName", remoteName,
		"ipRangeRemoteFallbackName", remoteFallbackName,
	)

	logger.Info("Loading GCP Address (V2)")

	addr, err := state.computeClient.GetIpRange(ctx, project, remoteName)

	if gcpmeta.IsNotFound(err) {
		// fallback to old name (backwards compatibility)
		logger.Info("New IpRange not found, checking the old name")
		fallbackAddr, err2 := state.computeClient.GetIpRange(ctx, project, remoteFallbackName)

		if gcpmeta.IsNotFound(err2) {
			logger.Info("Fallback IpRange name not found, proceeding")
			return nil, nil
		}

		if err2 != nil {
			logger.Error(err2, "Error getting fallback ipRange Addresses from GCP")
			return composed.UpdateStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: "Error getting fallback ipRange Addresses from GCP",
				}).
				SuccessError(composed.StopWithRequeue).
				SuccessLogMsg("Updated condition for failed IpRange fetching").
				Run(ctx, state)
		}

		if !strings.HasSuffix(fallbackAddr.Network, fmt.Sprintf("/%s", vpcName)) {
			logger.Info("Target fallback ipRange doesnt belong to this VPC, skipping")
			return nil, nil
		}

		state.address = fallbackAddr
		return nil, nil
	}

	if err != nil {
		logger.Error(err, "Error getting fallback ipRange Addresses from GCP")
		return composed.UpdateStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "Error getting Addresses from GCP",
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Updated condition for failed IpRange fetching").
			Run(ctx, state)
	}

	if !strings.HasSuffix(addr.Network, fmt.Sprintf("/%s", vpcName)) {
		logger.Error(errors.New("obtained IpRange belongs to another VPC"), "Obtained IpRange belongs to another VPC")
		return composed.UpdateStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "Obtained IpRange belongs to another VPC.",
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("Obtained IpRange belongs to another VPC").
			Run(ctx, state)
	}

	state.address = addr

	return nil, nil
}
