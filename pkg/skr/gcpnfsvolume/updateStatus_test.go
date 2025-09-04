package gcpnfsvolume

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type updateStatusSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *updateStatusSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *updateStatusSuite) TestWhenKcpStatusIsReady() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Remove Ready Status from GcpNfsVolume
	nfsVol := gcpNfsVolume.DeepCopy()
	nfsProtocol := "Any NFS Protocol"
	nfsVol.Status.Conditions = nil
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVol)
	assert.Nil(s.T(), err)

	state := factory.newStateWith(nfsVol)
	state.KcpNfsInstance = gcpNfsInstance.DeepCopy()
	state.KcpNfsInstance.SetStateData(client.GcpNfsStateDataProtocol, nfsProtocol)
	//Invoke updateStatus
	err, _ = updateStatus(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), err, composed.StopWithRequeue)

	//Get the modified GcpNfsVolume object
	nfsVol = &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(s.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(s.T(), gcpNfsInstance.Status.Conditions[0].Status, nfsVol.Status.Conditions[0].Status)
	assert.Equal(s.T(), gcpNfsInstance.Status.Conditions[0].Type, nfsVol.Status.Conditions[0].Type)
	assert.Equal(s.T(), cloudresourcesv1beta1.GcpNfsVolumeReady, nfsVol.Status.State)

	assert.Equal(s.T(), nfsProtocol, nfsVol.Status.Protocol)
}

func (s *updateStatusSuite) TestWhenKcpNSkrStatusAreReady() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state := factory.newState()
	state.KcpNfsInstance = &gcpNfsInstance

	//Invoke updateStatus
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(s.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(s.T(), metav1.ConditionTrue, nfsVol.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudresourcesv1beta1.ConditionTypeReady, nfsVol.Status.Conditions[0].Type)
	assert.Equal(s.T(), gcpNfsVolume.Status.State, nfsVol.Status.State)
	assert.Empty(s.T(), nfsVol.Status.Protocol, "Protocol should be empty when not set in KCP NfsInstance")
}

func (s *updateStatusSuite) TestWhenKcpStatusIsError() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

	//Remove Ready Status from GcpNfsVolume
	nfsVol := gcpNfsVolume.DeepCopy()
	nfsVol.Status.Conditions = nil
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVol)
	assert.Nil(s.T(), err)

	state := factory.newStateWith(nfsVol)
	state.KcpNfsInstance = nfsInstance

	//Invoke updateStatus
	err, _ = updateStatus(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), err, composed.StopAndForget)

	//Get the modified GcpNfsVolume object
	nfsVol = &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(s.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(s.T(), nfsInstance.Status.Conditions[0].Status, nfsVol.Status.Conditions[0].Status)
	assert.Equal(s.T(), nfsInstance.Status.Conditions[0].Type, nfsVol.Status.Conditions[0].Type)
	assert.Equal(s.T(), cloudresourcesv1beta1.GcpNfsVolumeError, nfsVol.Status.State)
	assert.Empty(s.T(), nfsVol.Status.Protocol, "Protocol should not be set when KCP NfsInstance is in error state")

}

func (s *updateStatusSuite) TestWhenKcpNSkrStatusAreError() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

	state := factory.newStateWith(nfsVol)
	state.KcpNfsInstance = nfsInstance

	//Invoke updateStatus
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), composed.StopAndForget, err)
	assert.Nil(s.T(), _ctx)

	//Get the modified GcpNfsVolume object
	nfsVol = &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(s.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(s.T(), metav1.ConditionTrue, nfsVol.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudresourcesv1beta1.ConditionTypeError, nfsVol.Status.Conditions[0].Type)
	assert.Equal(s.T(), "SKR NFS Volume is in error", nfsVol.Status.Conditions[0].Message)

}

func (s *updateStatusSuite) TestWhenKcpNSkrConditionsEmpty() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Update KCP NfsInstance with empty conditions
	nfsInstance := gcpNfsInstance.DeepCopy()
	nfsInstance.Status.Conditions = nil
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, nfsInstance)
	assert.Nil(s.T(), err)

	//Update SKR GcpNfsVolume with empty conditions
	nfsVol := gcpNfsVolume.DeepCopy()
	nfsVol.Status.Conditions = nil
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVol)
	assert.Nil(s.T(), err)

	state := factory.newStateWith(nfsVol)
	state.KcpNfsInstance = nfsInstance

	//Invoke updateStatus
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), composed.StopWithRequeueDelay(200*time.Millisecond), err)
	assert.NotNil(s.T(), _ctx)

	//Get the modified GcpNfsVolume object
	nfsVol = &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)
	assert.Nil(s.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), 0, len(nfsVol.Status.Conditions))
}

func TestUpdateStatus(t *testing.T) {
	suite.Run(t, new(updateStatusSuite))
}
