package awsnfsvolumerestore

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type loadSkrAwsNfsVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadSkrAwsNfsVolumeSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeOnDeletingObj() {

	deletingObj := deletingAwsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	s.Nil(err)

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolume(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeWhenExists() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	state.skrAwsNfsVolumeBackup = awsNfsVolumeBackup.DeepCopy()
	s.Nil(err)

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolume(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeWhenNotExists() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	state.skrAwsNfsVolumeBackup = awsNfsVolumeBackup.DeepCopy()
	s.Nil(err)

	//delete awsNfsvolume from k8s
	err = factory.skrCluster.K8sClient().Delete(context.Background(), awsNfsVolume.DeepCopy())
	s.Nil(err)
	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolume(ctx, state)
	s.Equal(composed.StopAndForget, err)
	s.Equal(ctx, _ctx)
}

func TestLoadSkrAwsNfsVolume(t *testing.T) {
	suite.Run(t, new(loadSkrAwsNfsVolumeSuite))
}
