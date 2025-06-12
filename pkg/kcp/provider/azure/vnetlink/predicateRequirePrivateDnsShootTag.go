package vnetlink

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func predicateRequireVNetShootTag(ctx context.Context, st composed.State) bool {
	state := st.(*State)

	if state.vnetLink != nil {
		// the VirtualNetworkLink is already created
		return false
	}

	return true
}
