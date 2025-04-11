package gcpnfsvolumerestore

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

type releaseLeaseSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *releaseLeaseSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *releaseLeaseSuite) TestReleaseLease_OtherLeased() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	// Being in deletion shouldn't have any impact on the logic
	obj := deletingGcpNfsVolumeRestore.DeepCopy()
	obj.DeletionTimestamp = &metav1.Time{Time: time.Now()}
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
	err, _ = releaseLease(ctx, state)
	assert.Nil(suite.T(), err)
	lease := &v1.Lease{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: fmt.Sprintf("restore-%s", state.GcpNfsVolume.Name), Namespace: state.GcpNfsVolume.Namespace}, lease)
	assert.Nil(suite.T(), err)
	suite.Equal(*lease.Spec.HolderIdentity, otherOwner)
}

func (suite *releaseLeaseSuite) TestReleaseLease_SelfLeased() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	// Being in deletion shouldn't have any impact on the logic
	obj := deletingGcpNfsVolumeRestore.DeepCopy()
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
	owner := fmt.Sprintf("%s/%s", obj.Namespace, obj.Name)
	err = factory.skrCluster.K8sClient().Create(ctx, &v1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("restore-%s", state.GcpNfsVolume.Name),
			Namespace: state.GcpNfsVolume.Namespace,
		},
		Spec: v1.LeaseSpec{
			HolderIdentity:       &owner,
			LeaseDurationSeconds: leaseDuration,
			AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			RenewTime:            &metav1.MicroTime{Time: time.Now()},
		},
	})
	assert.Nil(suite.T(), err)
	err, _ = releaseLease(ctx, state)
	assert.Nil(suite.T(), err)
	lease := &v1.Lease{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: fmt.Sprintf("restore-%s", state.GcpNfsVolume.Name), Namespace: state.GcpNfsVolume.Namespace}, lease)
	suite.True(apierrors.IsNotFound(err))
}

func TestReleaseLease(t *testing.T) {
	suite.Run(t, new(releaseLeaseSuite))
}
