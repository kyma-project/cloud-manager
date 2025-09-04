package backupschedule

import (
	"context"
	"testing"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
)

type checkSuspensionSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *checkSuspensionSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *checkSuspensionSuite) TestWhenNfsScheduleIsDeleting() {
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := checkSuspension(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *checkSuspensionSuite) TestWhenNotSuspended() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := checkSuspension(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *checkSuspensionSuite) TestWhenSuspended() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Suspend = true
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = checkSuspension(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), v1beta1.JobStateSuspended, fromK8s.Status.State)
}

func (s *checkSuspensionSuite) TestValidCronExpression() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = "* * * * *"
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = checkSuspension(ctx, state)

	//validate expected return values
	s.Nil(err)
}

func TestCheckSuspensionSuite(t *testing.T) {
	suite.Run(t, new(checkSuspensionSuite))
}
