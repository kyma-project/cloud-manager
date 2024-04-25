package gcpnfsvolume

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

type createPersistenceVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *createPersistenceVolumeSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *createPersistenceVolumeSuite) TestWhenNfsVolumeReady() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()

	err, _ctx := createPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), err, composed.StopWithRequeueDelay(3*time.Second))
	assert.Nil(suite.T(), _ctx)

	//Get the created PV object
	pvName := gcpNfsVolume.Name
	pv := v1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, &pv)

	//validate NFS attributes of PV
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), pv.Spec.NFS.Server, gcpNfsVolume.Status.Hosts[0])
	assert.Equal(suite.T(), pv.Spec.NFS.Path, fmt.Sprintf("/%s", gcpNfsVolume.Spec.FileShareName))

	//Validate PV Capacity
	expectedCapacity := int64(gcpNfsVolume.Status.CapacityGb) * 1024 * 1024 * 1024
	quantity, _ := pv.Spec.Capacity["storage"]
	pvQuantity, _ := quantity.AsInt64()
	assert.Equal(suite.T(), expectedCapacity, pvQuantity)
}

func (suite *createPersistenceVolumeSuite) TestWhenNfsVolumeDeleting() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&deletedGcpNfsVolume)

	err, _ctx := createPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Get the PV object
	pvName := fmt.Sprintf("%s--%s", deletedGcpNfsVolume.Namespace, deletedGcpNfsVolume.Name)
	pv := v1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, &pv)

	//validate for PV not found
	assert.NotNil(suite.T(), err)
	if status, ok := err.(*errors.StatusError); ok {
		assert.Equal(suite.T(), int32(404), status.ErrStatus.Code)
	}
}

func (suite *createPersistenceVolumeSuite) TestWhenGcpNfsVolumeNotReady() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	nfsVolume := cloudresourcesv1beta1.GcpNfsVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "not-ready-nfs-volume",
			Namespace: "test",
		},
	}
	state := factory.newStateWith(&nfsVolume)

	err, _ctx := createPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Get the PV object
	pvName := fmt.Sprintf("%s--%s", nfsVolume.Namespace, nfsVolume.Name)
	pv := v1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, &pv)

	//validate for PV not found
	assert.NotNil(suite.T(), err)
	if status, ok := err.(*errors.StatusError); ok {
		assert.Equal(suite.T(), int32(404), status.ErrStatus.Code)
	}
}

func (suite *createPersistenceVolumeSuite) TestWhenPVAlreadyPresent() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()
	state.PV = &v1.PersistentVolume{}

	err, _ctx := createPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Get the PV object
	pvName := fmt.Sprintf("%s--%s", gcpNfsVolume.Namespace, gcpNfsVolume.Name)
	pv := v1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, &pv)

	//validate for PV not found
	assert.NotNil(suite.T(), err)
	if status, ok := err.(*errors.StatusError); ok {
		assert.Equal(suite.T(), int32(404), status.ErrStatus.Code)
	}
}

func TestCreatePersistenceVolume(t *testing.T) {
	suite.Run(t, new(createPersistenceVolumeSuite))
}
