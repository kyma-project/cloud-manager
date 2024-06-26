package nfsbackupschedule

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

type loadNfsVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadNfsVolumeSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *loadNfsVolumeSuite) TestWhenNfsScheduleIsDeleting() {

	obj := deletingBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &gcpScope

	//Invoke loadNfsVolume API
	err, _ctx := loadNfsVolume(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadNfsVolumeSuite) TestWhenNfsScheduleExists() {

	obj := nfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &gcpScope

	err, _ctx := loadNfsVolume(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadNfsVolumeSuite) TestVolumeNotFound() {

	objDiffName := nfsBackupSchedule.DeepCopy()
	objDiffName.Spec.NfsVolumeRef.Name = "diffName"

	factory, err := newTestStateFactoryWithObj(objDiffName)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(objDiffName)
	suite.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &gcpScope

	//Invoke loadNfsVolume API
	err, _ctx := loadNfsVolume(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)
	suite.Equal(ctx, _ctx)
}

func (suite *loadNfsVolumeSuite) TestGcpVolumeNotReady() {

	obj := nfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &gcpScope

	// Remove the conditions from volume
	notReadyVolume := gcpNfsVolume.DeepCopy()
	notReadyVolume.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, notReadyVolume)
	suite.Nil(err)
	err, _ = loadNfsVolume(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)
	fromK8s := &v1beta1.NfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: nfsBackupSchedule.Name,
			Namespace: nfsBackupSchedule.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.JobStateError, fromK8s.Status.State)
	assert.Equal(suite.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (suite *loadNfsVolumeSuite) TestGcpVolumeReady() {

	obj := nfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &gcpScope

	//Invoke loadNfsVolume API
	err, ctx = loadNfsVolume(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
}

func (suite *loadNfsVolumeSuite) TestAwsVolumeNotReady() {

	obj := nfsBackupSchedule.DeepCopy()
	obj.Spec.NfsVolumeRef.Name = awsNfsVolume.Name
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &awsScope

	// Remove the conditions from volume
	notReadyVolume := awsNfsVolume.DeepCopy()
	notReadyVolume.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, notReadyVolume)
	suite.Nil(err)
	err, _ = loadNfsVolume(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)
	fromK8s := &v1beta1.NfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: nfsBackupSchedule.Name,
			Namespace: nfsBackupSchedule.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.JobStateError, fromK8s.Status.State)
	assert.Equal(suite.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (suite *loadNfsVolumeSuite) TestAwsVolumeReady() {

	obj := nfsBackupSchedule.DeepCopy()
	obj.Spec.NfsVolumeRef.Name = awsNfsVolume.Name
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &awsScope

	//Invoke loadNfsVolume API
	err, ctx = loadNfsVolume(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
}

func TestLoadNfsVolumeSuite(t *testing.T) {
	suite.Run(t, new(loadNfsVolumeSuite))
}
