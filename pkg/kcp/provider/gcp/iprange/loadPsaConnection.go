package iprange

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

const PsaPeeringName = "servicenetworking-googleapis-com"

// loadPsaConnection loads the Private Service Access (PSA) connection from GCP.
// PSA connections are used to connect VPCs to Google-managed services (like Cloud SQL, Redis).
// This action only runs for IpRanges configured with GcpPurposePSA.
func loadPsaConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()

	// Skip if this IpRange is not for PSA
	if ipRange.Spec.Options.Gcp != nil &&
		ipRange.Spec.Options.Gcp.Purpose != v1beta1.GcpPurposePSA {
		logger.Info("IpRange is not for PSA, skipping PSA connection load")
		return nil, nil
	}

	logger = logger.WithValues("ipRange", ipRange.Name)
	logger.Info("Loading GCP PSA Connection")

	// Get GCP scope details
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	// List all service connections for this VPC
	connections, err := state.serviceNetworkingClient.ListServiceConnections(ctx, project, vpc)
	if err != nil {
		logger.Error(err, "Error listing Service Connections from GCP")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "Error listing Service Connections from GCP",
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Updated condition for failed PSA connection listing").
			Run(ctx, state)
	}

	// Find the PSA connection (identified by specific peering name)
	for _, conn := range connections {
		if conn.Peering == PsaPeeringName {
			state.serviceConnection = conn
			logger.Info("Found PSA connection", "peering", conn.Peering)
			break
		}
	}

	if state.serviceConnection == nil {
		logger.Info("No PSA connection found")
	}

	return nil, nil
}
