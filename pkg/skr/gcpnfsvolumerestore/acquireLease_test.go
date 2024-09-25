package gcpnfsvolumerestore

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

type acquireLeaseSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *acquireLeaseSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *acquireLeaseSuite) TestAcquireLease_Acquire() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	assert.Nil(suite.T(), err)
	err, _ = acquireLease(ctx, state)
	assert.Nil(suite.T(), err)
	lease := &v1.Lease{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: fmt.Sprintf("restore-%s", state.GcpNfsVolume.Name), Namespace: state.GcpNfsVolume.Namespace}, lease)
	assert.Nil(suite.T(), err)
	suite.Equal(*lease.Spec.HolderIdentity, fmt.Sprintf("%s/%s", obj.Namespace, obj.Name))
}

func (suite *acquireLeaseSuite) TestAcquireLease_Renew() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	err, _ = acquireLease(ctx, state)
	assert.Nil(suite.T(), err)
	lease := &v1.Lease{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: fmt.Sprintf("restore-%s", state.GcpNfsVolume.Name), Namespace: state.GcpNfsVolume.Namespace}, lease)
	assert.Nil(suite.T(), err)
	suite.Equal(*lease.Spec.HolderIdentity, fmt.Sprintf("%s/%s", obj.Namespace, obj.Name))
	time1 := lease.Spec.RenewTime.Time
	err, _ = acquireLease(ctx, state)
	assert.Nil(suite.T(), err)
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: fmt.Sprintf("restore-%s", state.GcpNfsVolume.Name), Namespace: state.GcpNfsVolume.Namespace}, lease)
	assert.Nil(suite.T(), err)
	time2 := lease.Spec.RenewTime.Time
	suite.True(time2.After(time1))
}

func (suite *acquireLeaseSuite) TestAcquireLease_OtherLeased() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	leaseDuration := new(int32)
	*leaseDuration = 600
	otherOwner := "otherns/other"
	err = factory.skrCluster.K8sClient().Create(ctx, &v1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("restore-%s", state.GcpNfsVolume.Name),
			Namespace: state.GcpNfsVolume.Namespace,
		},
		Spec: v1.LeaseSpec{
			HolderIdentity:       &otherOwner,
			LeaseDurationSeconds: leaseDuration,
			AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			RenewTime:            &metav1.MicroTime{Time: time.Now()},
		},
	})
	assert.Nil(suite.T(), err)
	err, _ = acquireLease(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(util.Timing.T10000ms()), err)
	lease := &v1.Lease{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: fmt.Sprintf("restore-%s", state.GcpNfsVolume.Name), Namespace: state.GcpNfsVolume.Namespace}, lease)
	assert.Nil(suite.T(), err)
	suite.Equal(*lease.Spec.HolderIdentity, otherOwner)
}

func (suite *acquireLeaseSuite) TestDoNotAcquireLeaseOnDeletingObject() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := deletingGcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()

	err, _ = acquireLease(ctx, state)
	assert.Nil(suite.T(), err)
	lease := &v1.Lease{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: fmt.Sprintf("restore-%s", state.GcpNfsVolume.Name), Namespace: state.GcpNfsVolume.Namespace}, lease)
	assert.NotNil(suite.T(), err)
	assert.True(suite.T(), apierrors.IsNotFound(err))
}

func TestAcquireLease(t *testing.T) {
	suite.Run(t, new(acquireLeaseSuite))
}
