package gcpnfsvolume

import (
	"context"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

type updateStatusSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *updateStatusSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *updateStatusSuite) TestWhenKcpStatusIsReady() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Remove Ready Status from GcpNfsVolume
	nfsVol := gcpNfsVolume.DeepCopy()
	nfsVol.Status.Conditions = nil
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVol)
	assert.Nil(suite.T(), err)

	state := factory.newStateWith(nfsVol)
	state.KcpNfsInstance = &gcpNfsInstance

	//Invoke updateStatus
	err, _ = updateStatus(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), err, composed.StopWithRequeue)

	//Get the modified GcpNfsVolume object
	nfsVol = &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(suite.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(suite.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(suite.T(), gcpNfsInstance.Status.Conditions[0].Status, nfsVol.Status.Conditions[0].Status)
	assert.Equal(suite.T(), gcpNfsInstance.Status.Conditions[0].Type, nfsVol.Status.Conditions[0].Type)
}

func (suite *updateStatusSuite) TestWhenKcpNSkrStatusAreReady() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state := factory.newState()
	state.KcpNfsInstance = &gcpNfsInstance

	//Invoke updateStatus
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(suite.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(suite.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(suite.T(), metav1.ConditionTrue, nfsVol.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudresourcesv1beta1.ConditionTypeReady, nfsVol.Status.Conditions[0].Type)
}

func (suite *updateStatusSuite) TestWhenKcpStatusIsError() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Update KCP NfsInstance with ErrorStatus
	nfsInstance := gcpNfsInstance.DeepCopy()
	nfsInstance.Status.Conditions = []metav1.Condition{
		{
			Type:    "Error",
			Status:  "True",
			Reason:  "Error",
			Message: "NFS instance is not ready",
		},
	}
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, nfsInstance)
	assert.Nil(suite.T(), err)

	//Remove Ready Status from GcpNfsVolume
	nfsVol := gcpNfsVolume.DeepCopy()
	nfsVol.Status.Conditions = nil
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVol)
	assert.Nil(suite.T(), err)

	state := factory.newStateWith(nfsVol)
	state.KcpNfsInstance = nfsInstance

	//Invoke updateStatus
	err, _ = updateStatus(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), err, composed.StopAndForget)

	//Get the modified GcpNfsVolume object
	nfsVol = &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(suite.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(suite.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(suite.T(), nfsInstance.Status.Conditions[0].Status, nfsVol.Status.Conditions[0].Status)
	assert.Equal(suite.T(), nfsInstance.Status.Conditions[0].Type, nfsVol.Status.Conditions[0].Type)
}

func (suite *updateStatusSuite) TestWhenKcpNSkrStatusAreError() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Update KCP NfsInstance with ErrorStatus
	nfsInstance := gcpNfsInstance.DeepCopy()
	nfsInstance.Status.Conditions = []metav1.Condition{
		{
			Type:    "Error",
			Status:  "True",
			Reason:  "Error",
			Message: "NFS instance is not ready",
		},
	}
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, nfsInstance)
	assert.Nil(suite.T(), err)

	//Update SKR GcpNfsVolume Status to error
	nfsVol := gcpNfsVolume.DeepCopy()
	nfsVol.Status.Conditions = []metav1.Condition{
		{
			Type:    "Error",
			Status:  "True",
			Reason:  "Error",
			Message: "SKR NFS Volume is in error",
		},
	}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVol)
	assert.Nil(suite.T(), err)

	state := factory.newStateWith(nfsVol)
	state.KcpNfsInstance = nfsInstance

	//Invoke updateStatus
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), composed.StopAndForget, err)
	assert.Nil(suite.T(), _ctx)

	//Get the modified GcpNfsVolume object
	nfsVol = &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(suite.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(suite.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(suite.T(), metav1.ConditionTrue, nfsVol.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudresourcesv1beta1.ConditionTypeError, nfsVol.Status.Conditions[0].Type)
	assert.Equal(suite.T(), "SKR NFS Volume is in error", nfsVol.Status.Conditions[0].Message)

}

func (suite *updateStatusSuite) TestWhenKcpNSkrConditionsEmpty() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Update KCP NfsInstance with empty conditions
	nfsInstance := gcpNfsInstance.DeepCopy()
	nfsInstance.Status.Conditions = nil
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, nfsInstance)
	assert.Nil(suite.T(), err)

	//Update SKR GcpNfsVolume with empty conditions
	nfsVol := gcpNfsVolume.DeepCopy()
	nfsVol.Status.Conditions = nil
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVol)
	assert.Nil(suite.T(), err)

	state := factory.newStateWith(nfsVol)
	state.KcpNfsInstance = nfsInstance

	//Invoke updateStatus
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(200*time.Millisecond), err)
	assert.Nil(suite.T(), _ctx)

	//Get the modified GcpNfsVolume object
	nfsVol = &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)
	assert.Nil(suite.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(suite.T(), 0, len(nfsVol.Status.Conditions))
}

func TestUpdateStatus(t *testing.T) {
	suite.Run(t, new(updateStatusSuite))
}
