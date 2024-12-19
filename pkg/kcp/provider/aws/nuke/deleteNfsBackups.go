package nuke

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func deleteNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, rks := range state.ProviderResources {
		if rks.Kind == "AwsNfsVolumeBackup" && rks.Provider == cloudcontrolv1beta1.ProviderAws {
			for _, obj := range rks.Objects {
				backup := obj.(AwsBackup)
				if backup.Status == types.RecoveryPointStatusDeleting {
					continue
				}
				vault := ptr.Deref(backup.BackupVaultName, "")
				arn := ptr.Deref(backup.RecoveryPointArn, "")
				_, err := state.awsClient.DeleteRecoveryPoint(ctx, vault, arn)
				if err != nil {
					logger.Error(err, fmt.Sprintf("Error requesting Aws NfsVolume Backup deletion %s", backup.GetId()))
				}
			}
		}

	}
	return nil, nil
}
