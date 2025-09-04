package awsnfsvolumerestore

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type loadSkrAwsNfsVolumeBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadSkrAwsNfsVolumeBackupSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *loadSkrAwsNfsVolumeBackupSuite) TestLoadSkrAwsNfsVolumeBackupOnDeletingObj() {

	deletingObj := deletingAwsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	s.Nil(err)

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolumeBackup(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *loadSkrAwsNfsVolumeBackupSuite) TestLoadSkrAwsNfsVolumeBackupWhenExists() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	state.skrAwsNfsVolumeBackup = awsNfsVolumeBackup.DeepCopy()
	s.Nil(err)

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolumeBackup(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *loadSkrAwsNfsVolumeBackupSuite) TestLoadSkrAwsNfsVolumeBackupWhenNotExists() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	state.skrAwsNfsVolumeBackup = awsNfsVolumeBackup.DeepCopy()
	s.Nil(err)

	//delete awsNfsvolume from k8s
	err = factory.skrCluster.K8sClient().Delete(context.Background(), awsNfsVolumeBackup.DeepCopy())
	s.Nil(err)
	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolumeBackup(ctx, state)
	s.Equal(composed.StopAndForget, err)
	s.Equal(ctx, _ctx)
}

func TestLoadSkrAwsNfsVolumeBackup(t *testing.T) {
	suite.Run(t, new(loadSkrAwsNfsVolumeBackupSuite))
}
