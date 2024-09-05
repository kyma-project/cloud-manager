package redisinstance

import (
	"context"

	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func modifyTransitEncryptionEnabled(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()

	if state.elastiCacheReplicationGroup == nil {
		return composed.StopWithRequeue, nil
	}

	currentTransitEncryptionEnabled := ptr.Deref(state.elastiCacheReplicationGroup.TransitEncryptionEnabled, false)
	desiredTransitEncryptionEnabled := redisInstance.Spec.Instance.Aws.TransitEncryptionEnabled

	// when disabling transient encryption, we cant go from enabled to disabled
	// we must go from enabled to preferred and then to disabled
	isDisablingMidstep := !desiredTransitEncryptionEnabled &&
		state.elastiCacheReplicationGroup.TransitEncryptionMode == elasticacheTypes.TransitEncryptionModeRequired

	// when enabling transient encryption, we cant go from disabled to enabled
	// we must go from disabled to preferred and then to enabled
	isEnablingMidstep := desiredTransitEncryptionEnabled &&
		state.elastiCacheReplicationGroup.TransitEncryptionMode == ""

	isMidstep := isDisablingMidstep || isEnablingMidstep

	if (currentTransitEncryptionEnabled == desiredTransitEncryptionEnabled) &&
		state.elastiCacheReplicationGroup.TransitEncryptionMode != elasticacheTypes.TransitEncryptionModePreferred &&
		!isMidstep {
		return nil, nil
	}

	state.UpdateTransitEncryptionEnabled(desiredTransitEncryptionEnabled, isMidstep)

	return nil, nil
}
