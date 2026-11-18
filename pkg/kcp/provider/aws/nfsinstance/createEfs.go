package nfsinstance

import (
	"context"
	"errors"
	"fmt"

	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createEfs(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	if state.efs != nil {
		return nil, ctx
	}

	out, err := state.awsClient.CreateFileSystem(
		ctx,
		state.Obj().GetName(), // CreationToken: stable per NfsInstance, makes create idempotent
		efstypes.PerformanceMode(state.ObjAsNfsInstance().Spec.Instance.Aws.PerformanceMode),
		efstypes.ThroughputMode(state.ObjAsNfsInstance().Spec.Instance.Aws.Throughput),
		[]efstypes.Tag{
			{
				Key:   new("Name"),
				Value: new(state.Obj().GetName()),
			},
			{
				Key:   ptr.To(common.TagCloudManagerName),
				Value: new(state.Name().String()),
			},
			{
				Key:   ptr.To(common.TagCloudManagerRemoteName),
				Value: new(state.ObjAsNfsInstance().Spec.RemoteRef.String()),
			},
			{
				Key:   ptr.To(common.TagScope),
				Value: new(state.ObjAsNfsInstance().Spec.Scope.Name),
			},
			{
				Key:   ptr.To(common.TagShoot),
				Value: new(state.Scope().Spec.ShootName),
			},
		},
	)

	if err != nil {
		// Idempotent create collided with an existing EFS (same CreationToken). Treat as
		// success-in-progress: the next loadEfs finds it by CreationToken (tier 1).
		if _, ok := errors.AsType[*efstypes.FileSystemAlreadyExists](err); ok {
			logger.Info("AWS EFS already exists for creation token, requeueing to load it")
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
		}

		logger.Error(err, "Error creating AWS EFS")
		return composed.NewStatusPatcherComposed(state.ObjAsNfsInstance()).
			SetConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingFileSystem,
				Message: fmt.Sprintf("Failed creating file system: %s", err),
			}).
			OnSuccess(composed.RequeueAfter(util.Timing.T60000ms())).
			OnFailure(
				composed.Log("Error updating NfsInstance status due failed creating efs"),
				composed.RequeueAfter(util.Timing.T60000ms()),
			).
			Run(ctx, state.Cluster().K8sClient())
	}

	logger = logger.WithValues("efsId", ptr.Deref(out.FileSystemId, ""))
	ctx = composed.LoggerIntoCtx(ctx, logger)
	logger.Info("AWS EFS created")

	// Record the id and requeue; the next reconcile's loadEfs loads the EFS normally.
	return composed.NewStatusPatcherComposed(state.ObjAsNfsInstance()).
		MutateStatus(func(nfsInstance *cloudcontrolv1beta1.NfsInstance) {
			nfsInstance.Status.Id = ptr.Deref(out.FileSystemId, "")
		}).
		OnSuccess(composed.RequeueAfter(util.Timing.T10000ms())).
		OnFailure(
			composed.Log("Error updating NfsInstance status with created file system id"),
			composed.RequeueAfter(util.Timing.T10000ms()),
		).
		Run(ctx, state.Cluster().K8sClient())
}
