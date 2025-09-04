package awsnfsvolumebackup

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type stopIfVolumeNotReadySuite struct {
	suite.Suite
	ctx context.Context
}

func (s *stopIfVolumeNotReadySuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *stopIfVolumeNotReadySuite) TestStopIfVolumeNotReadyOnDeletingObject() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	s.Nil(err)
	state.Obj().SetFinalizers([]string{})

	//Call stopIfVolumeNotReady
	err, _ctx := stopIfVolumeNotReady(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *stopIfVolumeNotReadySuite) TestStopIfVolumeNotReadyWhenNfsVolumeReady() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the state to ready
	nfsVolume := awsNfsVolume.DeepCopy()
	nfsVolume.Status.State = cloudresourcesv1beta1.StateReady
	meta.SetStatusCondition(nfsVolume.Conditions(), metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  "AfsNfsVolumeBackup is ready",
		Message: "AfsNfsVolumeBackup is ready",
	})
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVolume)
	s.Nil(err)
	state.skrAwsNfsVolume = nfsVolume

	err, _ctx := stopIfVolumeNotReady(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *stopIfVolumeNotReadySuite) TestStopIfVolumeNotReadyWhenNfsVolumeNotReady() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//set AwsNfsVolume in state
	nfsVolume := awsNfsVolume.DeepCopy()
	nfsVolume.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVolume)
	s.Nil(err)
	state.skrAwsNfsVolume = nfsVolume

	err, _ = stopIfVolumeNotReady(ctx, state)
	s.Equal(composed.StopAndForget, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)

	errCondition := meta.FindStatusCondition(fromK8s.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	s.NotNil(errCondition)
	s.Equal(metav1.ConditionTrue, errCondition.Status)
	s.Equal(cloudresourcesv1beta1.ReasonNfsVolumeNotReady, errCondition.Reason)
}

func TestStopIfVolumeNotReady(t *testing.T) {
	suite.Run(t, new(stopIfVolumeNotReadySuite))
}
