package gcpnfsvolume

import (
	"context"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

type deleteKcpNfsInstanceSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *deleteKcpNfsInstanceSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *deleteKcpNfsInstanceSuite) TestWhenNfsVolumeNotDeleting() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()

	//Call deleteKcpNfsInstance method.
	err, _ctx := deleteKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Validate the NfsInstance object is not deleted.
	nfsInstance := cloudcontrolv1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Status.Id,
			Namespace: kymaRef.Namespace},
		&nfsInstance)
	assert.Nil(suite.T(), err)
	assert.True(suite.T(), nfsInstance.DeletionTimestamp.IsZero())
}

func (suite *deleteKcpNfsInstanceSuite) TestWhenNfsVolumeIsDeleting() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&deletedGcpNfsVolume)
	state.KcpNfsInstance = &gcpNfsInstanceToDelete

	//Call deleteKcpNfsInstance method.
	err, _ctx := deleteKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(3*time.Second), err)
	assert.Nil(suite.T(), _ctx)

	//Validate the NfsInstance object is not deleted.
	nfsInstance := cloudcontrolv1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: deletedGcpNfsVolume.Status.Id,
			Namespace: kymaRef.Namespace},
		&nfsInstance)
	assert.Nil(suite.T(), err)
	assert.False(suite.T(), nfsInstance.DeletionTimestamp.IsZero())
}

func (suite *deleteKcpNfsInstanceSuite) TestWhenKcpNfsInstanceDoNotExist() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	nfsVol := cloudresourcesv1beta1.GcpNfsVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "not-matching-gcp-nfs-volume",
			Namespace: "test",
			DeletionTimestamp: &metav1.Time{
				Time: time.Now(),
			},
		},
		Status: cloudresourcesv1beta1.GcpNfsVolumeStatus{
			Id: "not-matching-gcp-nfs-instance",
		},
	}
	state := factory.newStateWith(&nfsVol)

	//Call deleteKcpNfsInstance method.
	err, _ctx := deleteKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)
}

func TestDeleteKcpNfsInstanceSuite(t *testing.T) {
	suite.Run(t, new(deleteKcpNfsInstanceSuite))
}
