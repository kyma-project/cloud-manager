package mock2

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/stretchr/testify/require"
)

func TestE2EFilestoreBackup(t *testing.T) {

	t.Run("Filestore backups can be created, updated, listed and deleted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		s := newE2ETestSuite(ctx, t)

		location := "us-east1"

		net := s.createNetworkOK("test-net")
		addr := s.createPsaRangeOK(net.GetSelfLink(), "test-address", "10.251.0.0", 16)
		_ = s.createPsaConnectionOK(net.GetSelfLink(), addr.GetSelfLink())
		fs := s.createFilestoreOK(gcputil.NewLocationName(s.mock.ProjectId(), location).String(), "test-instance", net.GetSelfLink(), addr.GetSelfLink(), 1024)

		parentNd := gcputil.NewLocationName(s.mock.ProjectId(), location)

		backup := s.createFilestoreBackupOK(parentNd.String(), "backup-one", fs.Name, fs.FileShares[0].Name, map[string]string{"foo": "foo"})
		backup = s.updateFilestoreBackupLabelsOK(backup.Name, map[string]string{"foo": "bar"})

		backup2 := s.createFilestoreBackupOK(parentNd.String(), "backup-two", fs.Name, fs.FileShares[0].Name, map[string]string{"foo": "foo"})

		it := s.mock.ListFilestoreBackups(s.ctx, &filestorepb.ListBackupsRequest{
			Parent: parentNd.String(),
		}).All()
		arr, err := IteratorToSlice(it)
		require.NoError(t, err)
		require.Len(t, arr, 2)

		s.deleteFilestoreBackupOK(backup.Name)
		s.deleteFilestoreBackupOK(backup2.Name)

		// delete

		s.deleteFilestoreOK(fs.Name)
		s.deleteAddressOK(addr.GetName())
		s.deleteNetworkOK(net.GetName())
	})

}
