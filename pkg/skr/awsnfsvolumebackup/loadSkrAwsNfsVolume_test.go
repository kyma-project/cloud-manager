package awsnfsvolumebackup

import (
	"context"
	"fmt"
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

type loadSkrAwsNfsVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadSkrAwsNfsVolumeSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeOnDeletingObj() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	s.Nil(err)

	state.vault = nil

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolume(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeWhenExists() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolume(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeWhenVaultIsNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	obj.Spec.Source.Volume.Name = fmt.Sprintf("%s-new", obj.Spec.Source.Volume.Name)
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = loadSkrAwsNfsVolume(ctx, state)
	s.Equal(composed.StopAndForget, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: awsNfsVolumeBackup.Name,
			Namespace: awsNfsVolumeBackup.Namespace},
		fromK8s)
	s.Nil(err)

	errCondition := meta.FindStatusCondition(fromK8s.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	s.NotNil(errCondition)
	s.Equal(metav1.ConditionTrue, errCondition.Status)
	s.Equal(cloudresourcesv1beta1.ConditionReasonMissingNfsVolume, errCondition.Reason)

}

func TestLoadSkrAwsNfsVolume(t *testing.T) {
	suite.Run(t, new(loadSkrAwsNfsVolumeSuite))
}
