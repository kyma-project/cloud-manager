package nuke

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, rks := range state.ProviderResources {
		if rks.Kind != kindFilestoreBackup || rks.Provider != cloudcontrolv1beta1.ProviderGCP {
			continue
		}
		for _, obj := range rks.Objects {
			backup := obj.(GcpBackup)
			if backup.GetState() == filestorepb.Backup_DELETING {
				continue
			}
			// backup.Name is already the canonical full resource path.
			_, err := state.fileBackupClient.DeleteFilestoreBackup(ctx, &filestorepb.DeleteBackupRequest{
				Name: backup.GetName(),
			})
			if err != nil {
				logger.Error(err, fmt.Sprintf("Error requesting Gcp Filestore Backup deletion %s", backup.GetId()))
			}
		}
	}
	return nil, ctx
}
