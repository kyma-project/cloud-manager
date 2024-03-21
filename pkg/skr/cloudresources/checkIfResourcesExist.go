package cloudresources

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"time"
)

type checkItem struct {
	kind     string
	provider cloudcontrolv1beta1.ProviderType
	list     composed.ObjectList
}

func checkIfResourcesExist(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	var foundKinds []string
	checkItems := []*checkItem{
		// AWS =============================
		{
			kind:     "AwsNfsVolume",
			provider: cloudcontrolv1beta1.ProviderAws,
			list:     &cloudresourcesv1beta1.AwsNfsVolumeList{},
		},
		{
			kind:     "IpRange",
			provider: cloudcontrolv1beta1.ProviderAws,
			list:     &cloudresourcesv1beta1.IpRangeList{},
		},

		// GCP =============================
		{
			kind:     "GcpNfsVolume",
			provider: cloudcontrolv1beta1.ProviderGCP,
			list:     &cloudresourcesv1beta1.GcpNfsVolumeList{},
		},
		{
			kind:     "IpRange", // iprange can exist on all providers, so must be repeated here as well
			provider: cloudcontrolv1beta1.ProviderGCP,
			list:     &cloudresourcesv1beta1.IpRangeList{},
		},
	}

	for _, item := range checkItems {
		if _isProvider(state, item.provider) {
			if err := _testResourcesExist(ctx, state, item.list, item.kind, &foundKinds); err != nil {
				return composed.LogErrorAndReturn(err, fmt.Sprintf("Error listing %s", item.kind), composed.StopWithRequeue, ctx)
			}
		}
	}

	if len(foundKinds) == 0 {
		return nil, nil
	}

	logger.
		WithValues("existingResourceKinds", foundKinds).
		Info("Can not deactivate module due to found resources")

	state.ObjAsCloudResources().Status.State = cloudresourcesv1beta1.ModuleState(util.KymaModuleStateWarning)

	return composed.UpdateStatus(state.ObjAsCloudResources()).
		RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
		SuccessError(composed.StopWithRequeueDelay(300*time.Millisecond)).
		Run(ctx, state)
}

func _testResourcesExist(ctx context.Context, state *State, list composed.ObjectList, kind string, foundKinds *[]string) error {
	err := state.Cluster().K8sClient().List(ctx, list)
	if err != nil {
		return err
	}
	if list.GetItemCount() > 0 {
		*foundKinds = append(*foundKinds, kind)
	}
	return nil
}

func _isProvider(state *State, pt cloudcontrolv1beta1.ProviderType) bool {
	if state.Provider == nil {
		return true
	}
	if *state.Provider == pt {
		return true
	}
	return false
}
