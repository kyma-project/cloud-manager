package commonAction

import (
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// New returns a composed.Action that
//   - loads object being reconciled
//   - set stale status Ready Unknown Processing if generation != status.observedGeneration
//   - loads IpRange, GcpSubnet, VpcNetwork, Subscription so that:
//   - if not NotFound err, log and requeue
//   - if NotFound, set Ready False InvalidDependency with message "<KIND> <NAME> is not found" and requeue with commonrate.Slow10s
//   - if dependency implements composed.ObjWithStatus then it also
//   - if Ready condition does not exist or has status Unknown, requeue with commonrate.Slow1s
//   - if Ready condition has status False, set status Ready False InvalidDependency with message "<KIND> <NAME> is not ready" and requeue with commonrate.Slow10s
func New() composed.Action {
	return composed.ComposeActionsNoName(
		composed.LoadObj,
		statusStaleProcessing,
		ipRangeLoad,
		gcpSubnetLoad,
		vpcNetworkLoad,
		subscriptionLoad,
		// TODO: setup feature context
	)
}
