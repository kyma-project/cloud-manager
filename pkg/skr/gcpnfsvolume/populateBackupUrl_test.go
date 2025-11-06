package gcpnfsvolume

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type loadGcpNfsVolumeBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadGcpNfsVolumeBackupSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *loadGcpNfsVolumeBackupSuite) TestVolumeBackupNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	objDiffName := gcpNfsVolume.DeepCopy()
	objDiffName.Spec.SourceBackup.Name = "diffName"
	objDiffName.Spec.SourceBackup.Namespace = gcpNfsVolumeBackup.Namespace

	factory, err := newTestStateFactoryWithObject(&gcpNfsVolumeBackup, objDiffName)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state := factory.newStateWith(objDiffName)
	err, _ctx := populateBackupUrl(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(3*util.Timing.T1000ms()), err)
	s.Equal(ctx, _ctx)
}

func (s *loadGcpNfsVolumeBackupSuite) TestVolumeBackupNotReady() {
	obj := gcpNfsVolume.DeepCopy()
	obj.Spec.SourceBackup.Name = gcpNfsVolumeBackup.Name
	obj.Spec.SourceBackup.Namespace = gcpNfsVolumeBackup.Namespace
	backup := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObject(backup, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state := factory.newStateWith(obj)
	// Remove the conditions from backup
	backup.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, backup)
	s.Nil(err)

	err, _ = populateBackupUrl(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(3*util.Timing.T1000ms()), err)
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name,
			Namespace: gcpNfsVolume.Namespace},
		obj)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), v1beta1.GcpNfsVolumeError, obj.Status.State)
	assert.Equal(s.T(), metav1.ConditionTrue, obj.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudcontrolv1beta1.ConditionTypeError, obj.Status.Conditions[0].Type)
}

func (s *loadGcpNfsVolumeBackupSuite) TestVolumeBackupReady() {
	obj := gcpNfsVolume.DeepCopy()
	obj.Spec.SourceBackup.Name = gcpNfsVolumeBackup.Name
	obj.Spec.SourceBackup.Namespace = gcpNfsVolumeBackup.Namespace
	factory, err := newTestStateFactoryWithObject(&gcpNfsVolumeBackup, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state := factory.newStateWith(obj)
	state.Scope = kcpScope.DeepCopy()
	err, ctx = populateBackupUrl(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), ctx)
}

func (s *loadGcpNfsVolumeBackupSuite) TestVolumeBackupUrl() {
	obj := gcpNfsVolume.DeepCopy()
	obj.Spec.SourceBackupUrl = fmt.Sprintf("%s/%s", gcpNfsVolumeBackup.Status.Location, fmt.Sprintf("cm-%.60s", gcpNfsVolumeBackup.Status.Id))
	factory, err := newTestStateFactoryWithObject(&gcpNfsVolumeBackup, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state := factory.newStateWith(obj)
	state.Scope = kcpScope.DeepCopy()
	err, ctx = populateBackupUrl(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), ctx)
}

func TestLoadGcpNfsVolumeBackupSuite(t *testing.T) {
	suite.Run(t, new(loadGcpNfsVolumeBackupSuite))
}
