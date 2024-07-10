package gcpnfsvolume

import (
	"context"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type setProcessingSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *setProcessingSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *setProcessingSuite) TestSetProcessingWhenDeleting() {

	obj := deletedGcpNfsVolume.DeepCopy()
	factory, err := newTestStateFactory()
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(obj)
	suite.Nil(err)

	err, ctx = setProcessing(ctx, state)
	suite.Nil(err)
	suite.Nil(ctx)
}

func (suite *setProcessingSuite) TestSetProcessingWhenStateDone() {

	obj := gcpNfsVolume.DeepCopy()
	factory, err := newTestStateFactory()
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.StateReady
	meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  "test",
		Message: "test",
	})
	state := factory.newStateWith(obj)
	suite.Nil(err)

	err, ctx = setProcessing(ctx, state)
	suite.Nil(err)
	suite.Nil(ctx)
}

func (suite *setProcessingSuite) TestSetProcessingWhenStateError() {

	obj := gcpNfsVolume.DeepCopy()
	factory, err := newTestStateFactory()
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.StateError
	meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeError,
		Status:  metav1.ConditionTrue,
		Reason:  "test",
		Message: "test",
	})
	state := factory.newStateWith(obj)
	suite.Nil(err)

	err, ctx = setProcessing(ctx, state)
	suite.Nil(err)
	suite.Nil(ctx)
}

func (suite *setProcessingSuite) TestSetProcessingWhenStateEmpty() {

	obj := gcpNfsVolume.DeepCopy()
	factory, err := newTestStateFactory()
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Set the Status.State to empty.
	obj.Status.State = ""

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(obj)
	suite.Nil(err)

	err, ctx = setProcessing(ctx, state)
	suite.Equal(composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(v1beta1.GcpNfsVolumeProcessing, fromK8s.Status.State)
	suite.Nil(fromK8s.Status.Conditions)
}

func TestSetProcessing(t *testing.T) {
	suite.Run(t, new(setProcessingSuite))
}
