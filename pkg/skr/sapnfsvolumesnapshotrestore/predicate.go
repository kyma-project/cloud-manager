package sapnfsvolumesnapshotrestore

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func isInPlaceRestore(ctx context.Context, st composed.State) bool {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()
	return restore.Spec.Destination.ExistingVolume != nil
}
