package criprange

import (
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func New(factory StateFactory) composed.Action {
	return composed.ComposeActions(
		"crIpRangeMain",
		composed.LoadObj,
		validateCidr,
		copyCidrToStatus,
		preventCidrChange,
		addFinalizer,
		loadKcpIpRange,
		createKcpIpRange,
		deleteKcpIpRange,
		removeFinalizer,
		updateStatus,
		composed.StopAndForgetAction,
	)
}
