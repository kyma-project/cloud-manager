package gcpnfsvolume

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/http/httptest"
	"testing"
)

type loadGcpNfsVolumeBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadGcpNfsVolumeBackupSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *loadGcpNfsVolumeBackupSuite) TestVolumeBackupNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	objDiffName := gcpNfsVolume.DeepCopy()
	objDiffName.Spec.SourceBackup.Name = "diffName"
	objDiffName.Spec.SourceBackup.Namespace = gcpNfsVolumeBackup.Namespace

	factory, err := newTestStateFactoryWithObject(&gcpNfsVolumeBackup, objDiffName)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state := factory.newStateWith(objDiffName)
	err, _ctx := loadGcpNfsVolumeBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(3*util.Timing.T1000ms()), err)
	suite.Equal(ctx, _ctx)
}

func (suite *loadGcpNfsVolumeBackupSuite) TestVolumeBackupNotReady() {
	obj := gcpNfsVolume.DeepCopy()
	obj.Spec.SourceBackup.Name = gcpNfsVolumeBackup.Name
	obj.Spec.SourceBackup.Namespace = gcpNfsVolumeBackup.Namespace
	backup := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObject(backup, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state := factory.newStateWith(obj)
	// Remove the conditions from backup
	backup.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, backup)
	suite.Nil(err)

	err, _ = loadGcpNfsVolumeBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(3*util.Timing.T1000ms()), err)
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name,
			Namespace: gcpNfsVolume.Namespace},
		obj)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.GcpNfsVolumeError, obj.Status.State)
	assert.Equal(suite.T(), metav1.ConditionTrue, obj.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, obj.Status.Conditions[0].Type)
}

func (suite *loadGcpNfsVolumeBackupSuite) TestVolumeBackupReady() {
	obj := gcpNfsVolume.DeepCopy()
	obj.Spec.SourceBackup.Name = gcpNfsVolumeBackup.Name
	obj.Spec.SourceBackup.Namespace = gcpNfsVolumeBackup.Namespace
	factory, err := newTestStateFactoryWithObject(&gcpNfsVolumeBackup, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state := factory.newStateWith(obj)
	err, ctx = loadGcpNfsVolumeBackup(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
}

func TestLoadGcpNfsVolumeBackupSuite(t *testing.T) {
	suite.Run(t, new(loadGcpNfsVolumeBackupSuite))
}
