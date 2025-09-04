package gcpnfsvolume

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type loadKcpNfsInstanceSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadKcpNfsInstanceSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *loadKcpNfsInstanceSuite) TestWithMatchingNfsInstance() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()

	err, _ctx := loadKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Validate the NfsInstance object
	assert.NotNil(s.T(), state.KcpNfsInstance)
	assert.Equal(s.T(), gcpNfsVolume.Status.Id, state.KcpNfsInstance.Name)
}

func (s *loadKcpNfsInstanceSuite) TestWithNotMatchingNfsInstance() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	nfsVol := cloudresourcesv1beta1.GcpNfsVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "not-matching-gcp-nfs-volume",
			Namespace: "test",
		},
	}
	state := factory.newStateWith(&nfsVol)

	err, _ctx := loadKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Validate the NfsInstance object
	assert.Nil(s.T(), state.KcpNfsInstance)
}

func (s *loadKcpNfsInstanceSuite) TestWithMultipleMatchingNfsInstances() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get NfsInstance2 object from KcpCluster
	nfsInstance2 := cloudcontrolv1beta1.NfsInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gcp-nfs-volume-2",
			Namespace: kymaRef.Namespace,
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        kymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      "test-gcp-nfs-volume",
				cloudcontrolv1beta1.LabelRemoteNamespace: "test",
			},
		},
	}
	err = factory.kcpCluster.K8sClient().Create(ctx, &nfsInstance2)
	assert.Nil(s.T(), err)

	//Get state object with GcpNfsVolume
	state := factory.newState()

	err, _ctx := loadKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Validate the NfsInstance object
	assert.NotNil(s.T(), state.KcpNfsInstance)
}

func TestLoadKcpNfsInstanceSuite(t *testing.T) {
	suite.Run(t, new(loadKcpNfsInstanceSuite))
}
