package awsnfsvolumebackup

import (
	"context"
	"fmt"
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

type loadSkrAwsNfsVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadSkrAwsNfsVolumeSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeOnDeletingObj() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)

	state.vault = nil

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolume(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeWhenExists() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Call loadSkrAwsNfsVolume
	err, _ctx := loadSkrAwsNfsVolume(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadSkrAwsNfsVolumeSuite) TestLoadSkrAwsNfsVolumeWhenVaultIsNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	obj.Spec.Source.Volume.Name = fmt.Sprintf("%s-new", obj.Spec.Source.Volume.Name)
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = loadSkrAwsNfsVolume(ctx, state)
	suite.Equal(composed.StopAndForget, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: awsNfsVolumeBackup.Name,
			Namespace: awsNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err)

	errCondition := meta.FindStatusCondition(fromK8s.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	suite.NotNil(errCondition)
	suite.Equal(metav1.ConditionTrue, errCondition.Status)
	suite.Equal(cloudresourcesv1beta1.ConditionReasonMissingNfsVolume, errCondition.Reason)

}

func TestLoadSkrAwsNfsVolume(t *testing.T) {
	suite.Run(t, new(loadSkrAwsNfsVolumeSuite))
}
