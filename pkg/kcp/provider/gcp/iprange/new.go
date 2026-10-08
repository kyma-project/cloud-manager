package iprange

import (
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpiprangev3 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3"
)

// New returns the action for GCP IpRange provisioning via the v3 implementation.
func New(v3StateFactory gcpiprangev3.StateFactory) composed.Action {
	return gcpiprangev3.New(v3StateFactory)
}

// NewAllocateIpRangeAction returns the allocation action via the v3 implementation.
func NewAllocateIpRangeAction(v3StateFactory gcpiprangev3.StateFactory) composed.Action {
	return gcpiprangev3.NewAllocateIpRangeAction(v3StateFactory)
}
