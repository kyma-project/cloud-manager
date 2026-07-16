package nfsinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createFileSystem creates the NAS file system if it does not already exist, and persists
// its id to Status.Id so subsequent reconciliations can load it.
func createFileSystem(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.fileSystem != nil {
		return nil, ctx
	}

	alicloud := state.ObjAsNfsInstance().Spec.Instance.Alicloud
	protocolType := string(cloudcontrolv1beta1.AlicloudProtocolTypeNFS)
	storageType := string(cloudcontrolv1beta1.AlicloudStorageTypePerformance)
	if alicloud != nil {
		if alicloud.ProtocolType != "" {
			protocolType = string(alicloud.ProtocolType)
		}
		if alicloud.StorageType != "" {
			storageType = string(alicloud.StorageType)
		}
	}

	// Place the file system in the same zone as the first IpRange subnet.
	zoneId := ""
	if subnets := state.IpRange().Status.Subnets; len(subnets) > 0 {
		zoneId = subnets[0].Zone
	}

	logger.Info("Creating AliCloud NAS file system", "protocolType", protocolType, "storageType", storageType, "zoneId", zoneId)

	fsId, err := state.client.CreateFileSystem(ctx, protocolType, storageType, zoneId)
	if err != nil {
		logger.Error(err, "Error creating AliCloud NAS file system")
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		return composed.UpdateStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingFileSystem,
				Message: "Failed creating NAS file system",
			}).
			ErrorLogMessage("Error patching AliCloud KCP NfsInstance status after failed file system creation").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	state.fileSystemId = fsId
	state.ObjAsNfsInstance().Status.Id = fsId

	return composed.UpdateStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error patching AliCloud KCP NfsInstance status with file system id").
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
