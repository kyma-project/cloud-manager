package backupschedule

import (
	"context"
	"testing"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type validateTimeSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *validateTimeSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *validateTimeSuite) TestWhenNfsScheduleIsDeleting() {
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := validateTimes(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *validateTimeSuite) TestEmptyStartEndTimes() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := validateTimes(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *validateTimeSuite) TestInvalidStartTime() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.StartTime = &metav1.Time{Time: time.Now().AddDate(0, 0, -1)}
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = validateTimes(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)
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

func (s *validateTimeSuite) TestInvalidEndTime() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.EndTime = &metav1.Time{Time: time.Now().AddDate(0, 0, -1)}
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = validateTimes(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)
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
func (s *validateTimeSuite) TestValidStartAndInvalidEndTime() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.StartTime = &metav1.Time{Time: time.Now().AddDate(0, 0, 2)}
	obj.Spec.EndTime = &metav1.Time{Time: time.Now().AddDate(0, 0, 1)}
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = validateTimes(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)
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

func (s *validateTimeSuite) TestValidStartTime() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.StartTime = &metav1.Time{Time: time.Now().AddDate(0, 0, 1)}
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = validateSchedule(ctx, state)

	//validate expected return values
	s.Nil(err)
}

func (s *validateTimeSuite) TestValidEndTime() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.EndTime = &metav1.Time{Time: time.Now().AddDate(0, 0, 1)}
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = validateSchedule(ctx, state)

	//validate expected return values
	s.Nil(err)
}

func (s *validateTimeSuite) TestValidStartEndTime() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.StartTime = &metav1.Time{Time: time.Now().AddDate(0, 0, 1)}
	obj.Spec.EndTime = &metav1.Time{Time: time.Now().AddDate(0, 0, 2)}
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = validateSchedule(ctx, state)

	//validate expected return values
	s.Nil(err)
}

func TestValidateTimeSuite(t *testing.T) {
	suite.Run(t, new(validateTimeSuite))
}
