package gcpnfsvolume

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type checkRestorePermissionsSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *checkRestorePermissionsSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

// loadState seeds a backup with the given labels and runs loadScope + createBackupClient + loadBackup,
// returning the resulting state ready for a permission check.
func (s *checkRestorePermissionsSuite) loadState(labels map[string]string) *State {
	factory, err := newTestStateFactory(nil)
	assert.Nil(s.T(), err)

	err = seedBackup(factory, "us-west1", "backup-1", labels)
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

	return state
}

// Scenario (a): managed-by + GcpLabelShootName == shoot -> permitted.
func (s *checkRestorePermissionsSuite) TestOwnedByShootIsAllowed() {
	state := s.loadState(map[string]string{
		gcpclient.ManagedByKey: gcpclient.ManagedByValue,
		util.GcpLabelShootName: kcpScope.Spec.ShootName,
	})
	assert.True(s.T(), state.IsAllowedToRestoreBackup())
}

// Scenario (b): managed-by + cm-allow-<shoot> -> permitted.
func (s *checkRestorePermissionsSuite) TestAccessibleFromShootIsAllowed() {
	state := s.loadState(map[string]string{
		gcpclient.ManagedByKey:                              gcpclient.ManagedByValue,
		ConvertToAccessibleFromKey(kcpScope.Spec.ShootName): util.GcpLabelBackupAccessibleFrom,
	})
	assert.True(s.T(), state.IsAllowedToRestoreBackup())
}

// Scenario (c): managed-by only, neither owner nor accessible-from -> denied.
func (s *checkRestorePermissionsSuite) TestNeitherOwnedNorSharedIsDenied() {
	state := s.loadState(map[string]string{
		gcpclient.ManagedByKey: gcpclient.ManagedByValue,
		util.GcpLabelShootName: "some-other-shoot",
	})
	assert.False(s.T(), state.IsAllowedToRestoreBackup())
}

func TestCheckRestorePermissionsSuite(t *testing.T) {
	suite.Run(t, new(checkRestorePermissionsSuite))
}
