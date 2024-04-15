package gcpnfsvolumebackup

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func loadNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	backup := state.ObjAsGcpNfsVolumeBackup()
	logger.WithValues("Nfs Backup :", backup.Name).Info("Loading GCP File Instance")

	//Load the nfsInstance object
	nfsInstance := &cloudcontrolv1beta1.NfsInstance{}
	nfsInstanceKey := types.NamespacedName{
		Name:      state.GcpNfsVolume.Status.Id,
		Namespace: state.KymaRef.Namespace,
	}
	err := state.KcpCluster.K8sClient().Get(ctx, nfsInstanceKey, nfsInstance)
	if err != nil {
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: "Error loading NfsInstance from KCP",
			}).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Error getting Nfs Instance from GCP").
			Run(ctx, state)
	}

	//Store the nfsInstance in state
	state.NfsInstance = nfsInstance

	return nil, nil
}
