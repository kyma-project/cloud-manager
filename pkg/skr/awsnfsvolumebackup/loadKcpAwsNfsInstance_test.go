package awsnfsvolumebackup

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

type loadKcpAwsNfsInstanceSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadKcpAwsNfsInstanceSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *loadKcpAwsNfsInstanceSuite) TestLoadKcpAwsNfsInstanceOnDeletingObj() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)

	state.vault = nil

	//Call loadKcpAwsNfsInstance
	err, _ctx := loadKcpAwsNfsInstance(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadKcpAwsNfsInstanceSuite) TestLoadKcpAwsNfsInstanceWhenExists() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.skrAwsNfsVolume = awsNfsVolume.DeepCopy()

	//Call loadKcpAwsNfsInstance
	err, _ctx := loadKcpAwsNfsInstance(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadKcpAwsNfsInstanceSuite) TestLoadKcpAwsNfsInstanceWhenNotExists() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Update the KCP NfsInstance Id
	state.skrAwsNfsVolume = awsNfsVolume.DeepCopy()
	state.skrAwsNfsVolume.Status.Id =
		state.skrAwsNfsVolume.Status.Id + "_new"

	//Invoke API under test
	err, _ = loadKcpAwsNfsInstance(ctx, state)
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

func TestLoadKcpAwsNfsInstance(t *testing.T) {
	suite.Run(t, new(loadKcpAwsNfsInstanceSuite))
}
