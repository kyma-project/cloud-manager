package awsnfsvolumerestore

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type loadSkrAwsNfsVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadSkrAwsNfsVolumeSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeOnDeletingObj() {

	deletingObj := deletingAwsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolume(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeWhenExists() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	state.skrAwsNfsVolumeBackup = awsNfsVolumeBackup.DeepCopy()
	suite.Nil(err)

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolume(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeWhenNotExists() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	state.skrAwsNfsVolumeBackup = awsNfsVolumeBackup.DeepCopy()
	suite.Nil(err)

	//delete awsNfsvolume from k8s
	err = factory.skrCluster.K8sClient().Delete(context.Background(), awsNfsVolume.DeepCopy())
	suite.Nil(err)
	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolume(ctx, state)
	suite.Equal(composed.StopAndForget, err)
	suite.Equal(ctx, _ctx)
}

func TestLoadSkrAwsNfsVolume(t *testing.T) {
	suite.Run(t, new(loadSkrAwsNfsVolumeSuite))
}
