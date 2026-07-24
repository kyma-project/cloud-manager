package nuke

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmock2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/mock2"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const testProject = "test-project"
const testScopeName = "test-scope"

type nukeBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *nukeBackupSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

// newState builds a nuke gcp State backed by a mock2 store and a fake KCP client,
// with a GCP scope loaded so loadNfsBackups/deleteNfsBackup can run without envtest.
func (s *nukeBackupSuite) newState(nuke *cloudcontrolv1beta1.Nuke, store gcpmock2.Store) *State {
	kcpClient := fake.NewClientBuilder().
		WithScheme(commonscheme.KcpScheme).
		WithObjects(nuke).
		WithStatusSubresource(nuke).
		Build()
	cluster := composed.NewStateCluster(kcpClient, kcpClient, nil, commonscheme.KcpScheme)
	baseState := composed.NewStateFactory(cluster).NewState(
		types.NamespacedName{Name: nuke.Name, Namespace: nuke.Namespace}, nuke)
	focalState := focal.NewStateFactory().NewState(baseState)
	focalState.SetScope(&cloudcontrolv1beta1.Scope{
		ObjectMeta: metav1.ObjectMeta{Name: testScopeName},
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Gcp: &cloudcontrolv1beta1.GcpScope{Project: testProject},
			},
		},
	})
	return &State{
		State:            &nukeTestState{State: focalState},
		fileBackupClient: store,
	}
}

// nukeTestState satisfies nuketypes.State (focal.State + ObjAsNuke) for unit tests,
// mirroring pkg/kcp/nuke.State.ObjAsNuke without pulling in the reconciler wiring.
type nukeTestState struct {
	focal.State
}

func (s *nukeTestState) ObjAsNuke() *cloudcontrolv1beta1.Nuke {
	return s.Obj().(*cloudcontrolv1beta1.Nuke)
}

func (s *nukeBackupSuite) newStore() gcpmock2.Store {
	store := gcpmock2.New().NewSubscription(testProject)
	// NewSubscription generates a random project id; the state uses whatever the scope
	// declares, so seed/list operations here use the store directly and the reconciler
	// resolves the same store via the injected fileBackupClient.
	return store
}

func (s *nukeBackupSuite) seedBackup(store gcpmock2.Store, id string, state filestorepb.Backup_State) {
	err := store.AddFilestoreBackupDirectly(&filestorepb.Backup{
		Name:  gcpnfsbackupclientv2.GetFileBackupPath(testProject, "us-west1", id),
		State: state,
		Labels: map[string]string{
			gcpclient.ScopeNameKey: testScopeName,
			gcpclient.ManagedByKey: gcpclient.ManagedByValue,
		},
	})
	assert.NoError(s.T(), err)
}

func newNuke() *cloudcontrolv1beta1.Nuke {
	return &cloudcontrolv1beta1.Nuke{
		ObjectMeta: metav1.ObjectMeta{Name: "test-nuke", Namespace: "kcp-system"},
	}
}

// Scenario: deleteNfsBackup skips a backup already in state DELETING and requests
// deletion only for the others.
func (s *nukeBackupSuite) TestDeleteSkipsDeletingBackup() {
	store := s.newStore()
	s.seedBackup(store, "ready-backup", filestorepb.Backup_READY)
	s.seedBackup(store, "deleting-backup", filestorepb.Backup_DELETING)

	state := s.newState(newNuke(), store)

	err, _ := loadNfsBackups(s.ctx, state)
	assert.Nil(s.T(), err)

	err, _ = deleteNfsBackup(s.ctx, state)
	assert.Nil(s.T(), err)

	// The READY backup received a delete request (now DELETING); the pre-DELETING
	// backup was skipped (never re-requested). Resolve pending deletes and confirm
	// only the READY one is gone.
	assert.NoError(s.T(), store.ResolvePendingBackupDeleteOperations(s.ctx))

	iter := store.ListFilestoreBackups(s.ctx, &filestorepb.ListBackupsRequest{
		Parent: gcpnfsbackupclientv2.GetFilestoreParentPath(testProject, "-"),
		Filter: gcpclient.GetSkrBackupsFilter(testScopeName),
	})
	remaining, err := gcpmock2.IteratorToSlice(iter.All())
	assert.NoError(s.T(), err)
	assert.Len(s.T(), remaining, 1)
	assert.Equal(s.T(), filestorepb.Backup_DELETING, remaining[0].GetState())
}

// Scenario: loadNfsBackups reports an Error condition when the backup listing fails.
func (s *nukeBackupSuite) TestLoadReportsListingError() {
	store := s.newStore()
	store.SetListFilestoreBackupsError(errors.New("simulated GCP list backups failure"))

	nuke := newNuke()
	state := s.newState(nuke, store)

	err, _ := loadNfsBackups(s.ctx, state)
	// The action stops with a requeue error and patches the Nuke status.
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), string(cloudcontrolv1beta1.StateError), state.ObjAsNuke().Status.State)
	cond := meta.FindStatusCondition(state.ObjAsNuke().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(s.T(), cond)
	assert.Equal(s.T(), "ErrorListingGcpFilestoreBackups", cond.Reason)
}

func TestNukeBackupSuite(t *testing.T) {
	suite.Run(t, new(nukeBackupSuite))
}
