package nuke

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

func deleteNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, rks := range state.ProviderResources {
		if rks.Kind == "FilestoreBackup" && rks.Provider == cloudcontrolv1beta1.ProviderGCP {
			for _, obj := range rks.Objects {
				backup := obj.(GcpBackup)
				if backup.State == "DELETING" {
					continue
				}
				project, location, name := client.GetProjectLocationNameFromFileBackupPath(backup.Name)
				_, err := state.fileBackupClient.DeleteFileBackup(ctx, project, location, name)
				if err != nil {
					logger.Error(err, fmt.Sprintf("Error requesting Gcp Filestore Backup deletion %s", backup.GetId()))
				}
			}
		}

	}
	return nil, nil
}
