package gcpnfsvolume

import (
	"context"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

type deletePersistenceVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *deletePersistenceVolumeSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *deletePersistenceVolumeSuite) TestWhenNfsVolumeNotDeleting() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Create PV
	pv := pvGcpNfsVolume.DeepCopy()
	err = factory.skrCluster.K8sClient().Create(ctx, pv)
	assert.Nil(suite.T(), err)

	//Get state object with GcpNfsVolume
	state := factory.newState()
	state.PV = pv

	//Call deletePersistenceVolume.
	err, _ctx := deletePersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Validate the PV is not deleted.
	pv = &v1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: pvGcpNfsVolume.Name}, pv)
	assert.Nil(suite.T(), err)
	assert.True(suite.T(), pv.DeletionTimestamp.IsZero())
}

func (suite *deletePersistenceVolumeSuite) TestWhenNfsVolumeIsDeleting() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Create PV
	pv := pvDeletingGcpNfsVolume.DeepCopy()
	err = factory.skrCluster.K8sClient().Create(ctx, pv)
	assert.Nil(suite.T(), err)

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&deletedGcpNfsVolume)
	state.PV = &pvDeletingGcpNfsVolume

	//Call deletePersistenceVolume method.
	err, _ctx := deletePersistenceVolume(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(3*time.Second), err)
	assert.Nil(suite.T(), _ctx)

	//Validate the PV object is not deleted.
	pv = &v1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: pvDeletingGcpNfsVolume.Name}, pv)
	assert.Nil(suite.T(), err)
	assert.False(suite.T(), pv.DeletionTimestamp.IsZero())
}

func (suite *deletePersistenceVolumeSuite) TestWhenPVDoNotExist() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()

	//Call deletePersistenceVolume method.
	err, _ctx := deletePersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)
}

func (suite *deletePersistenceVolumeSuite) TestWhenPVHasWrongPhase() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Create PV
	pv := pvDeletingGcpNfsVolume.DeepCopy()
	pv.Status.Phase = "Bound"
	err = factory.skrCluster.K8sClient().Create(ctx, pv)
	assert.Nil(suite.T(), err)

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&deletedGcpNfsVolume)
	state.PV = pv

	//Call deletePersistenceVolume method.
	err, _ctx := deletePersistenceVolume(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), composed.StopAndForget, err)
	assert.Nil(suite.T(), _ctx)

	//Validate the Error Status is removed.
	nfsVolume := cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: deletedGcpNfsVolume.Name, Namespace: deletedGcpNfsVolume.Namespace},
		&nfsVolume)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, len(nfsVolume.Status.Conditions))
}

func (suite *deletePersistenceVolumeSuite) TestWhenPVBecomesReadyToDelete() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Create PV with Available Phase
	pv := pvDeletingGcpNfsVolume.DeepCopy()
	err = factory.skrCluster.K8sClient().Create(ctx, pv)
	assert.Nil(suite.T(), err)

	//Add Error Condition to GcpNfsVolume.
	nfsVolume := cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: deletedGcpNfsVolume.Name, Namespace: deletedGcpNfsVolume.Namespace},
		&nfsVolume)
	assert.Nil(suite.T(), err)
	nfsVolume.Status.Conditions = []metav1.Condition{
		{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ConditionReasonPVNotReadyForDeletion,
			Message: "test",
		},
	}

	//Update status
	err = factory.skrCluster.K8sClient().Status().Update(ctx, &nfsVolume)
	assert.Nil(suite.T(), err)

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&nfsVolume)
	state.PV = pv

	//Call deletePersistenceVolume method.
	err, _ctx := deletePersistenceVolume(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), composed.StopWithRequeue, err)
	assert.Nil(suite.T(), _ctx)

	//Validate the Error Status is removed.
	nfsVolume = cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: deletedGcpNfsVolume.Name, Namespace: deletedGcpNfsVolume.Namespace},
		&nfsVolume)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(nfsVolume.Status.Conditions))
}

func TestDeletePersistenceVolumeSuite(t *testing.T) {
	suite.Run(t, new(deletePersistenceVolumeSuite))
}
