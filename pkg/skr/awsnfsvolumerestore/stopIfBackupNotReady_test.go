package awsnfsvolumerestore

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

type stopIfBackupNotReadySuite struct {
	suite.Suite
	ctx context.Context
}

func (s *stopIfBackupNotReadySuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *stopIfBackupNotReadySuite) TestStopIfBackupNotReadyOnDeletingObject() {

	deletingObj := deletingAwsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	s.Nil(err)
	state.Obj().SetFinalizers([]string{})

	//Call stopIfBackupNotReady
	err, _ctx := stopIfBackupNotReady(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *stopIfBackupNotReadySuite) TestStopIfBackupNotReadyWhenNfsVolumeReady() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	state.skrAwsNfsVolumeBackup = awsNfsVolumeBackup.DeepCopy()

	err, _ctx := stopIfBackupNotReady(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *stopIfBackupNotReadySuite) TestStopIfBackupNotReadyWhenNfsVolumeNotReady() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//set AwsNfsVolume in state
	nfsVolumeBackup := awsNfsVolumeBackup.DeepCopy()
	nfsVolumeBackup.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVolumeBackup)
	s.Nil(err)
	state.skrAwsNfsVolumeBackup = nfsVolumeBackup

	err, _ = stopIfBackupNotReady(ctx, state)
	s.Equal(composed.StopAndForget, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)

	errCondition := meta.FindStatusCondition(fromK8s.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	s.NotNil(errCondition)
	s.Equal(metav1.ConditionTrue, errCondition.Status)
	s.Equal(cloudresourcesv1beta1.ConditionReasonNfsVolumeBackupNotReady, errCondition.Reason)
}

func TestStopIfBackupNotReady(t *testing.T) {
	suite.Run(t, new(stopIfBackupNotReadySuite))
}
