package nfsinstance

import (
	"context"
	"fmt"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
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
				Key:   pointer.String("Name"),
				Value: pointer.String(state.ObjAsNfsInstance().Spec.RemoteRef.String()),
			},
		},
	)

	if err != nil {
		logger.Error(err, "Error creating AWS EFS")
		meta.SetStatusCondition(state.ObjAsNfsInstance().Conditions(), metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed creating file system: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating NfsInstance status due failed creating efs", composed.StopWithRequeue, nil)
		}

		return composed.StopWithRequeueDelay(time.Minute), nil
	}

	logger.
		WithValues("efsId", out.FileSystemId).
		Info("AWS EFS created")

	//state.NfsInstance().Status.Id = pointer.StringDeref(out.FileSystemId, "")
	//meta.SetStatusCondition(state.NfsInstance().Conditions(), metav1.Condition{
	//	Type:               cloudresourcesv1beta1.ConditionTypeReady,
	//	Status:             "True",
	//	Reason:             cloudresourcesv1beta1.ReasonProvisioned,
	//	Message:            "AWS EFS is provisioned",
	//})
	//state.NfsInstance().Status.State = cloudresourcesv1beta1.ReadyState
	//
	//err = state.UpdateObjStatus(ctx)
	//if err != nil {
	//	return composed.LogErrorAndReturn(err, "Error updating NfsInstance status due to ready state", composed.StopWithRequeue, nil)
	//}

	return nil, nil
}
