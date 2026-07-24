package gcpnfsvolume

import (
	"context"
	"testing"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type loadBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadBackupSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

// seedBackup injects a filestore backup with the given labels into the factory's mock2 store,
// findable at the path derived from the test scope's project + the given location/name.
func seedBackup(f *testStateFactory, location, name string, labels map[string]string) error {
	path := gcpnfsbackupclientv2.GetFileBackupPath(kcpScope.Spec.Scope.Gcp.Project, location, name)
	return f.backupStore.AddFilestoreBackupDirectly(&filestorepb.Backup{
		Name:   path,
		Labels: labels,
		State:  filestorepb.Backup_READY,
	})
}

func newVolumeWithSourceBackupUrl(url string) *cloudresourcesv1beta1.GcpNfsVolume {
	vol := gcpNfsVolume.DeepCopy()
	vol.Spec.SourceBackupUrl = url
	return vol
}

// Scenario (d): backup lookup fails -> failed state + error condition, reconciliation stops.
func (s *loadBackupSuite) TestLoadBackupErrorSetsFailedState() {
	factory, err := newTestStateFactory(nil)
	assert.Nil(s.T(), err)

	// No backup seeded -> GetFilestoreBackup returns NotFound.
	vol := newVolumeWithSourceBackupUrl("us-west1/missing-backup")
	state, err := factory.newStateWith(vol)
	assert.Nil(s.T(), err)

	err, _ = loadScope(s.ctx, state)
	assert.Nil(s.T(), err)
	err, _ = createBackupClient(s.ctx, state)
	assert.Nil(s.T(), err)

	err, _ = loadBackup(s.ctx, state)
	// loadBackup patches status and returns StopAndForget on error.
	assert.Equal(s.T(), composed.StopAndForget, err)
	assert.Equal(s.T(), cloudresourcesv1beta1.GcpNfsVolumeState(cloudresourcesv1beta1.JobStateFailed), state.ObjAsGcpNfsVolume().Status.State)
	cond := meta.FindStatusCondition(state.ObjAsGcpNfsVolume().Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	assert.NotNil(s.T(), cond)
	assert.Equal(s.T(), cloudresourcesv1beta1.JobStateError, cond.Reason)
	assert.Nil(s.T(), state.fileBackup)
}

// Scenario (e): no SourceBackupUrl -> no lookup, no error.
func (s *loadBackupSuite) TestLoadBackupNoSourceBackupUrl() {
	factory, err := newTestStateFactory(nil)
	assert.Nil(s.T(), err)

	vol := newVolumeWithSourceBackupUrl("")
	state, err := factory.newStateWith(vol)
	assert.Nil(s.T(), err)

	err, _ = loadScope(s.ctx, state)
	assert.Nil(s.T(), err)
	err, _ = createBackupClient(s.ctx, state)
	assert.Nil(s.T(), err)

	err, _ = loadBackup(s.ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), state.fileBackup)
}

// Backup found -> stored on state as *filestorepb.Backup.
func (s *loadBackupSuite) TestLoadBackupFoundStoresBackup() {
	factory, err := newTestStateFactory(nil)
	assert.Nil(s.T(), err)

	err = seedBackup(factory, "us-west1", "backup-1", map[string]string{
		gcpclient.ManagedByKey: gcpclient.ManagedByValue,
		util.GcpLabelShootName: kcpScope.Spec.ShootName,
	})
	assert.Nil(s.T(), err)

	vol := newVolumeWithSourceBackupUrl("us-west1/backup-1")
	state, err := factory.newStateWith(vol)
	assert.Nil(s.T(), err)

	err, _ = loadScope(s.ctx, state)
	assert.Nil(s.T(), err)
	err, _ = createBackupClient(s.ctx, state)
	assert.Nil(s.T(), err)

	err, _ = loadBackup(s.ctx, state)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), state.fileBackup)
	assert.Equal(s.T(), gcpclient.ManagedByValue, state.fileBackup.GetLabels()[gcpclient.ManagedByKey])
}

func TestLoadBackupSuite(t *testing.T) {
	suite.Run(t, new(loadBackupSuite))
}
