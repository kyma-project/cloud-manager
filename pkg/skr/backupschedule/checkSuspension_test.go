package backupschedule

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

type checkSuspensionSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *checkSuspensionSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *checkSuspensionSuite) TestWhenNfsScheduleIsDeleting() {
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := checkSuspension(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *checkSuspensionSuite) TestWhenNotSuspended() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := checkSuspension(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *checkSuspensionSuite) TestWhenSuspended() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Suspend = true
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = checkSuspension(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopAndForget, err)
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.JobStateSuspended, fromK8s.Status.State)
}

func (suite *checkSuspensionSuite) TestValidCronExpression() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = "* * * * *"
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = checkSuspension(ctx, state)

	//validate expected return values
	suite.Nil(err)
}

func TestCheckSuspensionSuite(t *testing.T) {
	suite.Run(t, new(checkSuspensionSuite))
}
