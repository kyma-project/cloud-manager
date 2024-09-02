package awsnfsvolumebackup

import (
	"context"
	"fmt"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/suite"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type loadVaultSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadVaultSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *loadVaultSuite) TestLoadVaultOnDeletingObj() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)

	state.vault = nil

	//Call loadVault
	err, _ctx := loadVault(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadVaultSuite) TestLoadVaultWhenVaultIsNotNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	state.vault = &backuptypes.BackupVaultListMember{}

	//Call loadVault
	err, _ctx := loadVault(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadVaultSuite) TestLoadVaultWhenVaultIsNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//load the scope object into state
	awsScope := scope.DeepCopy()
	state.SetScope(awsScope)

	//Call createAwsClient
	err, _ = createAwsClient(ctx, state)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := loadVault(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)

	suite.NotNil(state.vault)
	suite.Equal(fmt.Sprintf("cm-%s", state.Scope().Name), ptr.Deref(state.vault.BackupVaultName, ""))
}

func TestLoadVault(t *testing.T) {
	suite.Run(t, new(loadVaultSuite))
}
