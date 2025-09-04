package awsnfsvolumebackup

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type shortCircuitSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *shortCircuitSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *shortCircuitSuite) TestWhenBackupIsDeleting() {

	obj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *shortCircuitSuite) TestWhenBackupIsReady() {

	config.AwsConfig.EfsCapacityCheckInterval = time.Hour * 6

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateReady
	obj.Status.Capacity = *resource.NewQuantity(int64(1024), resource.BinarySI)
	obj.Status.LastCapacityUpdate = &metav1.Time{Time: time.Now().Add(-1 * time.Hour)}
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	s.Equal(stopAndRequeueForCapacity(), err)
	s.Nil(_ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.StateReady, fromK8s.Status.State)
}

func (s *shortCircuitSuite) TestWhenBackupIsInError() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateError
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Equal(ctx, _ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.StateError, fromK8s.Status.State)
}

func (s *shortCircuitSuite) TestWhenBackupIsFailed() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateFailed
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)
	s.Nil(_ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.StateFailed, fromK8s.Status.State)
}

func (s *shortCircuitSuite) TestWhenBackupIsCreating() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateCreating
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Equal(ctx, _ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.StateCreating, fromK8s.Status.State)
}

func TestShortCircuit(t *testing.T) {
	suite.Run(t, new(shortCircuitSuite))
}
