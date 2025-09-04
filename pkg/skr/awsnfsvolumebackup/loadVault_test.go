package awsnfsvolumebackup

import (
	"context"
	"fmt"
	"testing"

	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/suite"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type loadVaultSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadVaultSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *loadVaultSuite) TestLoadVaultOnDeletingObj() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	s.Nil(err)

	state.vault = nil

	//Call loadVault
	err, _ctx := loadLocalVault(ctx, state)
	s.Nil(err)
	s.Equal(ctx, _ctx)
}

func (s *loadVaultSuite) TestLoadVaultWhenVaultIsNotNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	state.vault = &backuptypes.BackupVaultListMember{}

	//Call loadVault
	err, _ctx := loadLocalVault(ctx, state)
	s.Nil(err)
	s.Equal(ctx, _ctx)
}

func (s *loadVaultSuite) TestLoadVaultWhenVaultIsNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//load the scope object into state
	awsScope := scope.DeepCopy()
	state.SetScope(awsScope)

	//Call createAwsClient
	err, _ = createAwsClient(ctx, state)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := loadLocalVault(ctx, state)
	s.Nil(err)
	s.Equal(ctx, _ctx)

	s.NotNil(state.vault)
	s.Equal(fmt.Sprintf("cm-%s", state.Scope().Name), ptr.Deref(state.vault.BackupVaultName, ""))
}

func TestLoadVault(t *testing.T) {
	suite.Run(t, new(loadVaultSuite))
}
