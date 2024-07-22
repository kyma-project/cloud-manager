package nfsinstance

import (
	"context"
	"errors"
	"fmt"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"time"
)

func createEfs(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	if state.efs != nil {
		return nil, nil
	}

	out, err := state.awsClient.CreateFileSystem(
		ctx,
		efsTypes.PerformanceMode(state.ObjAsNfsInstance().Spec.Instance.Aws.PerformanceMode),
		efsTypes.ThroughputMode(state.ObjAsNfsInstance().Spec.Instance.Aws.Throughput),
		[]efsTypes.Tag{
			{
				Key:   ptr.To("Name"),
				Value: ptr.To(state.Obj().GetName()),
			},
			{
				Key:   ptr.To(common.TagCloudManagerName),
				Value: ptr.To(state.Name().String()),
			},
			{
				Key:   ptr.To(common.TagCloudManagerRemoteName),
				Value: ptr.To(state.ObjAsNfsInstance().Spec.RemoteRef.String()),
			},
			{
				Key:   ptr.To(common.TagScope),
				Value: ptr.To(state.ObjAsNfsInstance().Spec.Scope.Name),
			},
			{
				Key:   ptr.To(common.TagShoot),
				Value: ptr.To(state.Scope().Spec.ShootName),
			},
		},
	)

	if err != nil {
		logger.Error(err, "Error creating AWS EFS")
		meta.SetStatusCondition(state.ObjAsNfsInstance().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed creating file system: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error updating NfsInstance status due failed creating efs", ctx)
		}

		return composed.StopWithRequeueDelay(time.Minute), nil
	}

	logger = logger.WithValues("efsId", *out.FileSystemId)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	logger.Info("AWS EFS created")

	err, ctx = loadEfs(ctx, st)
	if err != nil {
		return err, ctx
	}

	if state.efs == nil {
		logger.Error(errors.New("unable to load just created EFS"), "Logical error!!!")
		return composed.UpdateStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonUnknown,
				Message: "Failed creating EFS",
			}).
			ErrorLogMessage("Error updating KCP NfsInstance status after failed loading of just created EFS").
			Run(ctx, state)
	}

	return nil, ctx
}
