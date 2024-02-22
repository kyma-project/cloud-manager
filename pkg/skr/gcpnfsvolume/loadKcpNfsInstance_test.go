package gcpnfsvolume

import (
	"context"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type loadKcpNfsInstanceSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadKcpNfsInstanceSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *loadKcpNfsInstanceSuite) TestWithMatchingNfsInstance() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()

	err, _ctx := loadKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Validate the NfsInstance object
	assert.NotNil(suite.T(), state.KcpNfsInstance)
	assert.Equal(suite.T(), gcpNfsVolume.Status.Id, state.KcpNfsInstance.Name)
}

func (suite *loadKcpNfsInstanceSuite) TestWithNotMatchingNfsInstance() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&deletedGcpNfsVolume)

	err, _ctx := loadKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Validate the NfsInstance object
	assert.Nil(suite.T(), state.KcpNfsInstance)
}

func (suite *loadKcpNfsInstanceSuite) TestWithMultipleMatchingNfsInstances() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get NfsInstance2 object from KcpCluster
	nfsInstance2 := cloudcontrolv1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx,
		types.NamespacedName{Namespace: gcpNfsInstance2.Namespace, Name: gcpNfsInstance2.Name},
		&nfsInstance2)
	assert.Nil(suite.T(), err)

	//Update it to have the matching labels as NfsInstance1
	nfsInstance2.Labels[cloudcontrolv1beta1.LabelRemoteName] = gcpNfsVolume.Name
	err = factory.kcpCluster.K8sClient().Update(ctx, &nfsInstance2)
	assert.Nil(suite.T(), err)

	//Get state object with GcpNfsVolume
	state := factory.newState()

	gcpNfsInstance2.Labels[cloudcontrolv1beta1.LabelRemoteName] = gcpNfsVolume.Name
	err, _ctx := loadKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Validate the NfsInstance object
	assert.NotNil(suite.T(), state.KcpNfsInstance)
}

func TestLoadKcpNfsInstanceSuite(t *testing.T) {
	suite.Run(t, new(loadKcpNfsInstanceSuite))
}
