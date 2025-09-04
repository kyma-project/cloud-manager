package gcpnfsvolume

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type createPersistenceVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *createPersistenceVolumeSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *createPersistenceVolumeSuite) TestWhenNfsVolumeReady() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()

	err, _ctx := createPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), err, composed.StopWithRequeueDelay(3*time.Second))
	assert.Nil(s.T(), _ctx)

	//Get the created PV object
	pvName := gcpNfsVolume.Status.Id
	pv := corev1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, &pv)

	//validate NFS attributes of PV
	assert.Nil(s.T(), err)
	assert.True(s.T(), len(pvName) > 0)
	assert.Equal(s.T(), pv.Spec.NFS.Server, gcpNfsVolume.Status.Hosts[0])
	assert.Equal(s.T(), pv.Spec.NFS.Path, fmt.Sprintf("/%s", gcpNfsVolume.Spec.FileShareName))

	//Validate PV Capacity
	expectedCapacity := gcpNfsVolume.Status.Capacity.Value()
	quantity := pv.Spec.Capacity["storage"]
	pvQuantity, _ := quantity.AsInt64()
	assert.Equal(s.T(), expectedCapacity, pvQuantity)
	assert.Empty(s.T(), pv.Spec.MountOptions)
}

func (s *createPersistenceVolumeSuite) TestWhenNfsVolumeReady41() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	volume := gcpNfsVolume.DeepCopy()
	volume.Status.Protocol = string(client.FilestoreProtocolNFSv41)
	state := factory.newStateWith(volume)

	err, _ctx := createPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), err, composed.StopWithRequeueDelay(3*time.Second))
	assert.Nil(s.T(), _ctx)

	//Get the created PV object
	pvName := gcpNfsVolume.Status.Id
	pv := corev1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, &pv)

	//validate NFS attributes of PV
	assert.Nil(s.T(), err)
	assert.True(s.T(), len(pvName) > 0)
	assert.Equal(s.T(), pv.Spec.NFS.Server, gcpNfsVolume.Status.Hosts[0])
	assert.Equal(s.T(), pv.Spec.NFS.Path, fmt.Sprintf("/%s", gcpNfsVolume.Spec.FileShareName))

	//Validate PV Capacity
	expectedCapacity := int64(gcpNfsVolume.Status.CapacityGb) * 1024 * 1024 * 1024
	quantity := pv.Spec.Capacity["storage"]
	pvQuantity, _ := quantity.AsInt64()
	assert.Equal(s.T(), expectedCapacity, pvQuantity)
	assert.NotEmpty(s.T(), pv.Spec.MountOptions, "should have nfsvers=4.1")
	assert.Equal(s.T(), "nfsvers=4.1", pv.Spec.MountOptions[0], "should have nfsvers=4.1 mount option")
}

func (s *createPersistenceVolumeSuite) TestWhenNfsVolumeDeleting() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&deletedGcpNfsVolume)

	err, _ctx := createPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Get the PV object
	pvName := fmt.Sprintf("%s--%s", deletedGcpNfsVolume.Namespace, deletedGcpNfsVolume.Name)
	pv := corev1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, &pv)

	//validate for PV not found
	assert.NotNil(s.T(), err)
	if status, ok := err.(*errors.StatusError); ok {
		assert.Equal(s.T(), int32(404), status.ErrStatus.Code)
	}
}

func (s *createPersistenceVolumeSuite) TestWhenGcpNfsVolumeNotReady() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Get the PV object
	pvName := fmt.Sprintf("%s--%s", nfsVolume.Namespace, nfsVolume.Name)
	pv := corev1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, &pv)

	//validate for PV not found
	assert.NotNil(s.T(), err)
	if status, ok := err.(*errors.StatusError); ok {
		assert.Equal(s.T(), int32(404), status.ErrStatus.Code)
	}
}

func (s *createPersistenceVolumeSuite) TestWhenPVAlreadyPresent() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()
	state.PV = &corev1.PersistentVolume{}

	err, _ctx := createPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Get the PV object
	pvName := fmt.Sprintf("%s--%s", gcpNfsVolume.Namespace, gcpNfsVolume.Name)
	pv := corev1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, &pv)

	//validate for PV not found
	assert.NotNil(s.T(), err)
	if status, ok := err.(*errors.StatusError); ok {
		assert.Equal(s.T(), int32(404), status.ErrStatus.Code)
	}
}

func TestCreatePersistenceVolume(t *testing.T) {
	suite.Run(t, new(createPersistenceVolumeSuite))
}
