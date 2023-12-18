package scope

import (
	"github.com/kyma-project/cloud-resources/components/lib/composed"
)

func WhenNoScope() composed.Action {
	return composed.ComposeActions(
		"whenNoScope",
		loadKyma,
		createGardenerClient,
		loadShoot,
		loadGardenerCredentials,
		createScope,
		saveScope,
		updateScopeRef,
		// scope is created, requeue now
		composed.StopWithRequeueAction,
	)
}
