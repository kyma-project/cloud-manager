package gcpnfsvolume

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type modifyPersistenceVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *modifyPersistenceVolumeSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *modifyPersistenceVolumeSuite) TestWhenNfsVolumeReady() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	nfsVol := cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, &nfsVol)
	assert.Nil(s.T(), err)
	state := factory.newStateWith(&nfsVol)

	//Create a PV.
	pv := pvGcpNfsVolume.DeepCopy()
	err = factory.skrCluster.K8sClient().Create(ctx, pv)
	assert.Nil(s.T(), err)
	state.PV = pv

	//Update the capacity in spec.
	nfsVol.Spec.CapacityGb = 2048
	err = factory.skrCluster.K8sClient().Update(ctx, &nfsVol)
	assert.Nil(s.T(), err)

	//Update the capacity in status
	nfsVol.Status.CapacityGb = nfsVol.Spec.CapacityGb
	nfsVol.Status.Capacity = resource.MustParse(fmt.Sprintf("%dGi", nfsVol.Spec.CapacityGb))
	err = factory.skrCluster.K8sClient().Status().Update(ctx, &nfsVol)
	assert.Nil(s.T(), err)

	//Invoke modifyPersistenceVolume
	err, _ctx := modifyPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), err, composed.StopWithRequeueDelay(1*time.Second))
	assert.Nil(s.T(), _ctx)

	//Get the modified PV object
	pvName := fmt.Sprintf("%s--%s", gcpNfsVolume.Namespace, gcpNfsVolume.Name)
	pv = &corev1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, pv)

	//validate NFS attributes of PV
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), pv.Spec.NFS.Server, gcpNfsVolume.Status.Hosts[0])
	assert.Equal(s.T(), pv.Spec.NFS.Path, fmt.Sprintf("/%s", gcpNfsVolume.Spec.FileShareName))

	//Validate PV Capacity
	expectedCapacity := nfsVol.Status.Capacity.Value()
	quantity := pv.Spec.Capacity["storage"]
	pvQuantity, _ := quantity.AsInt64()
	assert.Equal(s.T(), expectedCapacity, pvQuantity)
}

func (s *modifyPersistenceVolumeSuite) TestWhenNfsVolumeDeleting() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&deletedGcpNfsVolume)

	err, _ctx := modifyPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)
}

func (s *modifyPersistenceVolumeSuite) TestWhenNfsVolumeNotReady() {
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

	err, _ctx := modifyPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)
}

func (s *modifyPersistenceVolumeSuite) TestWhenPVNotPresent() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()

	err, _ctx := modifyPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)
}

func (s *modifyPersistenceVolumeSuite) TestWhenNfsVolumeNotChanged() {
	factory, err := newTestStateFactory()
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()

	err, _ctx := modifyPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)
}

func TestModifyPersistenceVolume(t *testing.T) {
	suite.Run(t, new(modifyPersistenceVolumeSuite))
}
