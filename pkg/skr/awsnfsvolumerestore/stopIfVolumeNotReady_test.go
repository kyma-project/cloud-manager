package awsnfsvolumerestore

import (
	"context"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type stopIfVolumeNotReadySuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *stopIfVolumeNotReadySuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *stopIfVolumeNotReadySuite) TestStopIfVolumeNotReadyOnDeletingObject() {

	deletingObj := deletingAwsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)
	state.Obj().SetFinalizers([]string{})

	//Call stopIfVolumeNotReady
	err, _ctx := stopIfVolumeNotReady(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *stopIfVolumeNotReadySuite) TestStopIfVolumeNotReadyWhenNfsVolumeReady() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the state to ready
	nfsVolume := awsNfsVolume.DeepCopy()
	nfsVolume.Status.State = cloudresourcesv1beta1.StateReady
	meta.SetStatusCondition(nfsVolume.Conditions(), metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  "AfsNfsVolume is ready",
		Message: "AfsNfsVolume is ready",
	})
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVolume)
	suite.Nil(err)
	state.skrAwsNfsVolume = nfsVolume

	err, _ctx := stopIfVolumeNotReady(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *stopIfVolumeNotReadySuite) TestStopIfVolumeNotReadyWhenNfsVolumeNotReady() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//set AwsNfsVolume in state
	nfsVolume := awsNfsVolume.DeepCopy()
	nfsVolume.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVolume)
	suite.Nil(err)
	state.skrAwsNfsVolume = nfsVolume

	err, _ = stopIfVolumeNotReady(ctx, state)
	suite.Equal(composed.StopAndForget, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	errCondition := meta.FindStatusCondition(fromK8s.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	suite.NotNil(errCondition)
	suite.Equal(metav1.ConditionTrue, errCondition.Status)
	suite.Equal(cloudresourcesv1beta1.ReasonNfsVolumeNotReady, errCondition.Reason)
}

func TestStopIfVolumeNotReady(t *testing.T) {
	suite.Run(t, new(stopIfVolumeNotReadySuite))
}
