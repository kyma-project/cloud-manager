package backupschedule

import (
	"context"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type loadSourceSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadSourceSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *loadSourceSuite) TestWhenScheduleIsDeleting() {

	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &gcpScope
	state.backupImpl = &backupImplGcpNfs{}

	//Invoke loadSource API
	err, _ctx := loadSource(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *loadSourceSuite) TestWhenScheduleExists() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &gcpScope
	state.backupImpl = &backupImplGcpNfs{}

	err, _ctx := loadSource(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *loadSourceSuite) TestSourceRefNotFound() {

	objDiffName := gcpNfsBackupSchedule.DeepCopy()
	objDiffName.Spec.NfsVolumeRef.Name = "diffName"

	factory, err := newTestStateFactoryWithObj(objDiffName)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(objDiffName)
	s.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &gcpScope
	state.backupImpl = &backupImplGcpNfs{}

	//Invoke loadSource API
	err, _ctx := loadSource(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(util.Timing.T10000ms()), err)
	s.Equal(ctx, _ctx)
}

func (s *loadSourceSuite) TestSourceRefNotReady() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &gcpScope
	state.backupImpl = &backupImplGcpNfs{}

	// Remove the conditions from volume
	notReadyVolume := gcpNfsVolume.DeepCopy()
	notReadyVolume.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, notReadyVolume)
	s.Nil(err)
	err, _ = loadSource(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(util.Timing.T10000ms()), err)
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), v1beta1.JobStateError, fromK8s.Status.State)
	assert.Equal(s.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (s *loadSourceSuite) TestSourceRefReady() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the gcpScope object in state
	state.Scope = &gcpScope
	state.backupImpl = &backupImplGcpNfs{}

	//Invoke loadSource API
	err, ctx = loadSource(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), ctx)
}

func TestLoadSourceSuite(t *testing.T) {
	suite.Run(t, new(loadSourceSuite))
}
