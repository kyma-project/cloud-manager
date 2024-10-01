package awsnfsvolumerestore

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type loadSkrAwsNfsVolumeBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadSkrAwsNfsVolumeBackupSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *loadSkrAwsNfsVolumeBackupSuite) TestLoadSkrAwsNfsVolumeBackupOnDeletingObj() {

	deletingObj := deletingAwsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolumeBackup(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadSkrAwsNfsVolumeBackupSuite) TestLoadSkrAwsNfsVolumeBackupWhenExists() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	state.skrAwsNfsVolumeBackup = awsNfsVolumeBackup.DeepCopy()
	suite.Nil(err)

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolumeBackup(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadSkrAwsNfsVolumeBackupSuite) TestLoadSkrAwsNfsVolumeBackupWhenNotExists() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	state.skrAwsNfsVolumeBackup = awsNfsVolumeBackup.DeepCopy()
	suite.Nil(err)

	//delete awsNfsvolume from k8s
	err = factory.skrCluster.K8sClient().Delete(context.Background(), awsNfsVolumeBackup.DeepCopy())
	suite.Nil(err)
	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolumeBackup(ctx, state)
	suite.Equal(composed.StopAndForget, err)
	suite.Equal(ctx, _ctx)
}

func TestLoadSkrAwsNfsVolumeBackup(t *testing.T) {
	suite.Run(t, new(loadSkrAwsNfsVolumeBackupSuite))
}
