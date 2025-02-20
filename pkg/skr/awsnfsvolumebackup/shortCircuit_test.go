package awsnfsvolumebackup

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type shortCircuitSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *shortCircuitSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *shortCircuitSuite) TestWhenBackupIsDeleting() {

	obj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *shortCircuitSuite) TestWhenBackupIsReady() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateReady
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopAndForget, err)
	suite.Nil(_ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.StateReady, fromK8s.Status.State)
}

func (suite *shortCircuitSuite) TestWhenBackupIsInError() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateError
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Equal(ctx, _ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.StateError, fromK8s.Status.State)
}

func (suite *shortCircuitSuite) TestWhenBackupIsFailed() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateFailed
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopAndForget, err)
	suite.Nil(_ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.StateFailed, fromK8s.Status.State)
}

func (suite *shortCircuitSuite) TestWhenBackupIsCreating() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateCreating
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Equal(ctx, _ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.StateCreating, fromK8s.Status.State)
}

func TestShortCircuit(t *testing.T) {
	suite.Run(t, new(shortCircuitSuite))
}
