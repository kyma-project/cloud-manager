package awsnfsvolumebackup

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

type markFailedSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *markFailedSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *markFailedSuite) TestWhenBackupIsDeleting() {

	obj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := markFailed(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *markFailedSuite) TestWhenBackupIsReady() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateReady
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := markFailed(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Equal(ctx, _ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.StateReady, fromK8s.Status.State)
}

func (suite *markFailedSuite) TestWhenBackupIsFailed() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateFailed
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := markFailed(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Equal(ctx, _ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.StateFailed, fromK8s.Status.State)
}

func (suite *markFailedSuite) TestWhenBackupIsCreating() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateCreating
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := markFailed(ctx, state)

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

func (suite *markFailedSuite) TestWhenBackupIsLatestAndInError() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateError
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := markFailed(ctx, state)

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

func (suite *markFailedSuite) TestWhenBackupIsNotLatestAndInError() {

	labels := map[string]string{
		v1beta1.LabelScheduleName:      "test-schedule",
		v1beta1.LabelScheduleNamespace: "test",
	}

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.CreationTimestamp = v1.Time{Time: time.Now().Add(-1 * time.Minute)}
	obj.Labels = labels
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	obj.Status.State = v1beta1.StateError
	err = state.Cluster().K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Create another backup object for the same schedule
	obj2 := awsNfsVolumeBackup.DeepCopy()
	obj2.Name = "test-backup-02"
	obj2.Namespace = "test"
	obj2.CreationTimestamp = v1.Time{Time: time.Now()}
	obj2.Labels = labels
	obj2.Status.State = v1beta1.StateReady
	err = factory.skrCluster.K8sClient().Create(ctx, obj2)
	suite.Nil(err)

	err, _ctx := markFailed(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopAndForget, err)
	suite.Equal(ctx, _ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.StateFailed, fromK8s.Status.State)
}

func TestMarkFailed(t *testing.T) {
	suite.Run(t, new(markFailedSuite))
}
