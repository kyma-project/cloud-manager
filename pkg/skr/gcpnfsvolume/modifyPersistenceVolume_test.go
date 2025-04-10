package gcpnfsvolume

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

type modifyPersistenceVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *modifyPersistenceVolumeSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *modifyPersistenceVolumeSuite) TestWhenNfsVolumeReady() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	nfsVol := cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, &nfsVol)
	assert.Nil(suite.T(), err)
	state := factory.newStateWith(&nfsVol)

	//Create a PV.
	pv := pvGcpNfsVolume.DeepCopy()
	err = factory.skrCluster.K8sClient().Create(ctx, pv)
	assert.Nil(suite.T(), err)
	state.PV = pv

	//Update the capacity in spec.
	nfsVol.Spec.CapacityGb = 2048
	err = factory.skrCluster.K8sClient().Update(ctx, &nfsVol)
	assert.Nil(suite.T(), err)

	//Update the capacity in status
	nfsVol.Status.CapacityGb = nfsVol.Spec.CapacityGb
	err = factory.skrCluster.K8sClient().Status().Update(ctx, &nfsVol)
	assert.Nil(suite.T(), err)

	//Invoke modifyPersistenceVolume
	err, _ctx := modifyPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), err, composed.StopWithRequeueDelay(1*time.Second))
	assert.Nil(suite.T(), _ctx)

	//Get the modified PV object
	pvName := fmt.Sprintf("%s--%s", gcpNfsVolume.Namespace, gcpNfsVolume.Name)
	pv = &corev1.PersistentVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, pv)

	//validate NFS attributes of PV
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), pv.Spec.NFS.Server, gcpNfsVolume.Status.Hosts[0])
	assert.Equal(suite.T(), pv.Spec.NFS.Path, fmt.Sprintf("/%s", gcpNfsVolume.Spec.FileShareName))

	//Validate PV Capacity
	expectedCapacity := int64(nfsVol.Status.CapacityGb) * 1024 * 1024 * 1024
	quantity := pv.Spec.Capacity["storage"]
	pvQuantity, _ := quantity.AsInt64()
	assert.Equal(suite.T(), expectedCapacity, pvQuantity)
}

func (suite *modifyPersistenceVolumeSuite) TestWhenNfsVolumeDeleting() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&deletedGcpNfsVolume)

	err, _ctx := modifyPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)
}

func (suite *modifyPersistenceVolumeSuite) TestWhenNfsVolumeNotReady() {
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

	err, _ctx := modifyPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)
}

func (suite *modifyPersistenceVolumeSuite) TestWhenPVNotPresent() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()

	err, _ctx := modifyPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)
}

func (suite *modifyPersistenceVolumeSuite) TestWhenNfsVolumeNotChanged() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newState()

	err, _ctx := modifyPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)
}

func TestModifyPersistenceVolume(t *testing.T) {
	suite.Run(t, new(modifyPersistenceVolumeSuite))
}
