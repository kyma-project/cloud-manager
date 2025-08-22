package awsnfsvolumebackup

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
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
		state.Scope().Spec.Region,
		awsconfig.AwsConfig.Default.AccessKeyId,
		awsconfig.AwsConfig.Default.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: err.Error(),
			}).
			SuccessLogMsg(fmt.Sprintf("Error assuming AWS role : %s", err)).
			SuccessError(err).
			Run(ctx, state)
	}

	state.awsClient = cli
	state.roleName = roleName

	//Create a remote AWS client (if needed)
	if !state.requiresRemoteBackup() {
		return nil, ctx
	}

	cli, err = state.awsClientProvider(
		ctx,
		state.Scope().Spec.Scope.Aws.AccountId,
		backup.Spec.Location,
		awsconfig.AwsConfig.Default.AccessKeyId,
		awsconfig.AwsConfig.Default.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: err.Error(),
			}).
			SuccessLogMsg(fmt.Sprintf("Error connecting to AWS remote region : %s", err)).
			SuccessError(err).
			Run(ctx, state)
	}
	state.destAwsClient = cli

	return nil, ctx
}
