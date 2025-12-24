package v3

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

// loadAddress loads the GCP global address resource for the IpRange.
// It supports backward compatibility by checking both new name format (cm-<uuid>)
// and old name format (spec.remoteRef.name).
func loadAddress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpcName := gcpScope.VpcNetwork

	// Try new name format first (cm-<uuid>)
	remoteName := GetIpRangeName(ipRange.GetName())
	// Fallback to old name format (spec.remoteRef.name) for backward compatibility
	remoteFallbackName := ipRange.Spec.RemoteRef.Name

	logger = logger.WithValues(
		"ipRange", ipRange.Name,
		"ipRangeRemoteName", remoteName,
		"ipRangeRemoteFallbackName", remoteFallbackName,
	)

	logger.Info("Loading GCP Address")

	// Try loading with new name format
	addr, err := state.computeClient.GetIpRange(ctx, project, remoteName)

	if gcpmeta.IsNotFound(err) {
		// Fallback to old name for backward compatibility
		logger.Info("New IpRange name not found, checking fallback name")
		fallbackAddr, err2 := state.computeClient.GetIpRange(ctx, project, remoteFallbackName)

		if gcpmeta.IsNotFound(err2) {
			// Neither name exists - resource not yet created
			logger.Info("Fallback IpRange name not found, proceeding with creation")
			return nil, nil
		}

		if err2 != nil {
			logger.Error(err2, "Error getting fallback IpRange address from GCP")
			return composed.PatchStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: "Error getting fallback IpRange address from GCP",
				}).
				SuccessError(composed.StopWithRequeue).
				SuccessLogMsg("Updated condition for failed IpRange fetching").
				Run(ctx, state)
		}

		// Defensive nil check
		if fallbackAddr == nil {
			logger.Error(errors.New("unexpected nil fallback address"), "Fallback address is nil despite no error")
			return composed.PatchStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: "Unexpected nil fallback address from GCP",
				}).
				SuccessError(composed.StopWithRequeue).
				SuccessLogMsg("Updated condition for unexpected nil fallback address").
				Run(ctx, state)
		}

		// Validate fallback address belongs to correct VPC
		if fallbackAddr.Network == nil || !strings.HasSuffix(*fallbackAddr.Network, fmt.Sprintf("/%s", vpcName)) {
			logger.Info("Fallback IpRange doesn't belong to this VPC, skipping")
			return nil, nil
		}

		state.address = fallbackAddr
		return nil, nil
	}

	if err != nil {
		logger.Error(err, "Error getting IpRange address from GCP")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "Error getting address from GCP",
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Updated condition for failed IpRange fetching").
			Run(ctx, state)
	}

	// Defensive nil check
	if addr == nil {
		logger.Error(errors.New("unexpected nil address"), "Address is nil despite no error")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "Unexpected nil address from GCP",
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Updated condition for unexpected nil address").
			Run(ctx, state)
	}

	// Validate address belongs to correct VPC
	if addr.Network == nil || !strings.HasSuffix(*addr.Network, fmt.Sprintf("/%s", vpcName)) {
		logger.Error(errors.New("obtained IpRange belongs to another VPC"), "Obtained IpRange belongs to another VPC")
		return composed.PatchStatus(ipRange).
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
