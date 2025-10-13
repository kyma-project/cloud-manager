package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func predicateRequireVNetShootTag(ctx context.Context, st composed.State) bool {
	state := st.(*State)

	if state.remotePeering != nil {
		// the peering is already created
		return false
	}

	if state.RemoteNetwork().Status.NetworkType == cloudcontrolv1beta1.NetworkTypeKyma {
		return false
	}
	if state.RemoteNetwork().Status.NetworkType == cloudcontrolv1beta1.NetworkTypeCloudResources {
		return false
	}

	return true
}
