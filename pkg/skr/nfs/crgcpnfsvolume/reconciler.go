package crgcpnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler struct {
}

func (r *Reconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := r.newState()
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *Reconciler) newState() *State {
	return nil
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crGcpNfsVolumeMain",
		composed.LoadObj,
		composed.ComposeActions(
			"crGcpNfsVolumeValidateSpec",
			validateIpRange, validateCapacity, validateFileShareName),
		addFinalizer,
		loadKcpNfsInstance,
		loadPersistenceVolume,
		modifyKcpNfsInstance,
		deletePersistenceVolume,
		deleteKcpNfsInstance,
		removeFinalizer,
		removePersistenceVolumeFinalizer,
		modifyPersistenceVolume,
		updateStatus,
		composed.StopAndForgetAction,
	)
}
