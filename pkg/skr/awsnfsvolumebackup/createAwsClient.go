package awsnfsvolumebackup

import (
	"context"
	"fmt"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createAwsClient(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAwsNfsVolumeBackup()

	roleName := fmt.Sprintf(
		"arn:aws:iam::%s:role/%s",
		state.Scope().Spec.Scope.Aws.AccountId,
		awsconfig.AwsConfig.Default.AssumeRoleName,
	)

	logger.
		WithValues(
			"awsRegion", state.Scope().Spec.Region,
			"awsRole", roleName,
		).
		Info("Assuming AWS role")

	cli, err := state.awsClientProvider(
		ctx,
		state.Scope().Spec.Scope.Aws.AccountId,
		state.GetBackupLocation(),
		awsconfig.AwsConfig.Default.AccessKeyId,
		awsconfig.AwsConfig.Default.SecretAccessKey,
		roleName,
	)
	if err != nil {
		backup.Status.State = cloudresourcesv1beta1.ConditionTypeError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeError,
				Message: err.Error(),
			}).
			SuccessLogMsg("AwsNfsVolumeBackup status got updated with error : "+err.Error()).
			SuccessError(composed.StopWithRequeueDelay(time.Second*3)).
			Run(ctx, state)
	}

	state.awsClient = cli
	state.roleName = roleName

	return nil, nil
}
