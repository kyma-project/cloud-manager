package gcpnfsvolume

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type modifyKcpNfsInstanceSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *modifyKcpNfsInstanceSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *modifyKcpNfsInstanceSuite) TestCreateNfsInstance() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newState()
	assert.Nil(s.T(), err)
	state.KcpIpRange = &kcpIpRange

	//Invoke modifyKcpNfsInstance
	err, _ = modifyKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), err, composed.StopWithRequeue)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), gcpNfsInstance.Name, gcpNfsVolume.Status.Id)

	//Get the KcpNfsInstance using theGcpNfsVolume.Status.Id
	nfsInstance := cloudcontrolv1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Status.Id, Namespace: kymaRef.Namespace}, &nfsInstance)
	assert.Nil(s.T(), err)

	//Validate KCP NfsInstance labels.
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelKymaName)
	assert.Equal(s.T(), kymaRef.Name, nfsInstance.Labels[cloudcontrolv1beta1.LabelKymaName])
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelRemoteName)
	assert.Equal(s.T(), gcpNfsVolume.Name, nfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteName])
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelRemoteNamespace)
	assert.Equal(s.T(), gcpNfsVolume.Namespace, nfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteNamespace])

	//Validate KCPNfsInstance attributes.
	assert.Equal(s.T(), kymaRef.Name, nfsInstance.Spec.Scope.Name)
	assert.Equal(s.T(), gcpNfsVolume.Name, nfsInstance.Spec.RemoteRef.Name)
	assert.Equal(s.T(), gcpNfsVolume.Namespace, nfsInstance.Spec.RemoteRef.Namespace)
	assert.Equal(s.T(), kcpIpRange.Name, nfsInstance.Spec.IpRange.Name)
	assert.Equal(s.T(), gcpNfsVolume.Spec.CapacityGb, nfsInstance.Spec.Instance.Gcp.CapacityGb)
	assert.Equal(s.T(), string(gcpNfsVolume.Spec.Tier), string(nfsInstance.Spec.Instance.Gcp.Tier))
	assert.Equal(s.T(), gcpNfsVolume.Spec.Location, nfsInstance.Spec.Instance.Gcp.Location)
	assert.Equal(s.T(), gcpNfsVolume.Spec.FileShareName, nfsInstance.Spec.Instance.Gcp.FileShareName)
	assert.Equal(s.T(), gcpNfsInstance.Spec.Instance.Gcp.ConnectMode, nfsInstance.Spec.Instance.Gcp.ConnectMode)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), cloudresourcesv1beta1.GcpNfsVolumeProcessing, nfsVol.Status.State)
}

func (s *modifyKcpNfsInstanceSuite) TestCreateNfsInstanceWithRestore() {
	obj := gcpNfsVolume.DeepCopy()
	obj.Spec.SourceBackup = cloudresourcesv1beta1.GcpNfsVolumeBackupRef{
		Name:      gcpNfsVolumeBackup.Name,
		Namespace: gcpNfsVolumeBackup.Namespace,
	}
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactoryWithObject(fakeHttpServer, &gcpNfsVolumeBackup, obj, &deletedGcpNfsVolume)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	s.Nil(err)
	state.KcpIpRange = &kcpIpRange
	srcBackupFullPath := gcpNfsVolumeBackupToUrl(&gcpNfsVolumeBackup)
	state.SrcBackupFullPath = srcBackupFullPath
	state.Scope = &kcpScope

	//Invoke modifyKcpNfsInstance
	err, _ = modifyKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), err, composed.StopWithRequeue)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), gcpNfsInstance.Name, nfsVol.Status.Id)

	//Get the KcpNfsInstance using theGcpNfsVolume.Status.Id
	nfsInstance := cloudcontrolv1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: nfsVol.Status.Id, Namespace: kymaRef.Namespace}, &nfsInstance)
	assert.Nil(s.T(), err)

	//Validate KCP NfsInstance labels.
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelKymaName)
	assert.Equal(s.T(), kymaRef.Name, nfsInstance.Labels[cloudcontrolv1beta1.LabelKymaName])
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelRemoteName)
	assert.Equal(s.T(), nfsVol.Name, nfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteName])
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelRemoteNamespace)
	assert.Equal(s.T(), nfsVol.Namespace, nfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteNamespace])

	//Validate KCPNfsInstance attributes.
	assert.Equal(s.T(), kymaRef.Name, nfsInstance.Spec.Scope.Name)
	assert.Equal(s.T(), gcpNfsVolume.Name, nfsInstance.Spec.RemoteRef.Name)
	assert.Equal(s.T(), gcpNfsVolume.Namespace, nfsInstance.Spec.RemoteRef.Namespace)
	assert.Equal(s.T(), kcpIpRange.Name, nfsInstance.Spec.IpRange.Name)
	assert.Equal(s.T(), gcpNfsVolume.Spec.CapacityGb, nfsInstance.Spec.Instance.Gcp.CapacityGb)
	assert.Equal(s.T(), string(gcpNfsVolume.Spec.Tier), string(nfsInstance.Spec.Instance.Gcp.Tier))
	assert.Equal(s.T(), gcpNfsVolume.Spec.Location, nfsInstance.Spec.Instance.Gcp.Location)
	assert.Equal(s.T(), gcpNfsVolume.Spec.FileShareName, nfsInstance.Spec.Instance.Gcp.FileShareName)
	assert.Equal(s.T(), gcpNfsInstance.Spec.Instance.Gcp.ConnectMode, nfsInstance.Spec.Instance.Gcp.ConnectMode)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), cloudresourcesv1beta1.GcpNfsVolumeProcessing, nfsVol.Status.State)
	assert.Equal(s.T(), "projects/test-project/locations/us-west1/backups/cm-backup-uuid", nfsInstance.Spec.Instance.Gcp.SourceBackup)
}

func (s *modifyKcpNfsInstanceSuite) TestCreateNfsInstanceWithRestoreBackupUrl() {
	obj := gcpNfsVolume.DeepCopy()
	obj.Spec.SourceBackupUrl = fmt.Sprintf("projects/%s/locations/%s/backups/%s", kcpScope.Spec.Scope.Gcp.Project, gcpNfsVolumeBackup.Status.Location, fmt.Sprintf("cm-%.60s", gcpNfsVolumeBackup.Status.Id))
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactoryWithObject(fakeHttpServer, &gcpNfsVolumeBackup, obj, &deletedGcpNfsVolume)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	s.Nil(err)
	state.KcpIpRange = &kcpIpRange
	srcBackupFullPath := gcpNfsVolumeBackupToUrl(&gcpNfsVolumeBackup)
	state.SrcBackupFullPath = srcBackupFullPath
	state.Scope = &kcpScope

	//Invoke modifyKcpNfsInstance
	err, _ = modifyKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), err, composed.StopWithRequeue)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), gcpNfsInstance.Name, nfsVol.Status.Id)

	//Get the KcpNfsInstance using theGcpNfsVolume.Status.Id
	nfsInstance := cloudcontrolv1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: nfsVol.Status.Id, Namespace: kymaRef.Namespace}, &nfsInstance)
	assert.Nil(s.T(), err)

	//Validate KCP NfsInstance labels.
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelKymaName)
	assert.Equal(s.T(), kymaRef.Name, nfsInstance.Labels[cloudcontrolv1beta1.LabelKymaName])
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelRemoteName)
	assert.Equal(s.T(), nfsVol.Name, nfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteName])
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelRemoteNamespace)
	assert.Equal(s.T(), nfsVol.Namespace, nfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteNamespace])

	//Validate KCPNfsInstance attributes.
	assert.Equal(s.T(), kymaRef.Name, nfsInstance.Spec.Scope.Name)
	assert.Equal(s.T(), gcpNfsVolume.Name, nfsInstance.Spec.RemoteRef.Name)
	assert.Equal(s.T(), gcpNfsVolume.Namespace, nfsInstance.Spec.RemoteRef.Namespace)
	assert.Equal(s.T(), kcpIpRange.Name, nfsInstance.Spec.IpRange.Name)
	assert.Equal(s.T(), gcpNfsVolume.Spec.CapacityGb, nfsInstance.Spec.Instance.Gcp.CapacityGb)
	assert.Equal(s.T(), string(gcpNfsVolume.Spec.Tier), string(nfsInstance.Spec.Instance.Gcp.Tier))
	assert.Equal(s.T(), gcpNfsVolume.Spec.Location, nfsInstance.Spec.Instance.Gcp.Location)
	assert.Equal(s.T(), gcpNfsVolume.Spec.FileShareName, nfsInstance.Spec.Instance.Gcp.FileShareName)
	assert.Equal(s.T(), gcpNfsInstance.Spec.Instance.Gcp.ConnectMode, nfsInstance.Spec.Instance.Gcp.ConnectMode)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), cloudresourcesv1beta1.GcpNfsVolumeProcessing, nfsVol.Status.State)
	assert.Equal(s.T(), "projects/test-project/locations/us-west1/backups/cm-backup-uuid", nfsInstance.Spec.Instance.Gcp.SourceBackup)
}

func (s *modifyKcpNfsInstanceSuite) TestCreateNfsInstanceNoLocation() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj := gcpNfsVolume.DeepCopy()
	obj.Spec.Location = ""
	state, err := factory.newStateWith(obj)
	s.Nil(err)
	state.KcpIpRange = &kcpIpRange
	state.Scope = &kcpScope
	//Invoke modifyKcpNfsInstance
	err, _ = modifyKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), err, composed.StopWithRequeue)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), nfsVol.Name, nfsVol.Status.Id)

	//Get the KcpNfsInstance using theGcpNfsVolume.Status.Id
	nfsInstance := cloudcontrolv1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Status.Id, Namespace: kymaRef.Namespace}, &nfsInstance)
	assert.Nil(s.T(), err)

	//Validate KCP NfsInstance labels.
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelKymaName)
	assert.Equal(s.T(), kymaRef.Name, nfsInstance.Labels[cloudcontrolv1beta1.LabelKymaName])
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelRemoteName)
	assert.Equal(s.T(), obj.Name, nfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteName])
	assert.Contains(s.T(), nfsInstance.Labels, cloudcontrolv1beta1.LabelRemoteNamespace)
	assert.Equal(s.T(), obj.Namespace, nfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteNamespace])

	//Validate KCPNfsInstance attributes.
	assert.Equal(s.T(), kymaRef.Name, nfsInstance.Spec.Scope.Name)
	assert.Equal(s.T(), obj.Name, nfsInstance.Spec.RemoteRef.Name)
	assert.Equal(s.T(), obj.Namespace, nfsInstance.Spec.RemoteRef.Namespace)
	assert.Equal(s.T(), kcpIpRange.Name, nfsInstance.Spec.IpRange.Name)
	assert.Equal(s.T(), obj.Spec.CapacityGb, nfsInstance.Spec.Instance.Gcp.CapacityGb)
	assert.Equal(s.T(), string(obj.Spec.Tier), string(nfsInstance.Spec.Instance.Gcp.Tier))
	assert.Contains(s.T(), kcpScope.Spec.Scope.Gcp.Workers[0].Zones, nfsInstance.Spec.Instance.Gcp.Location)
	assert.Equal(s.T(), obj.Spec.FileShareName, nfsInstance.Spec.Instance.Gcp.FileShareName)
	assert.Equal(s.T(), gcpNfsInstance.Spec.Instance.Gcp.ConnectMode, nfsInstance.Spec.Instance.Gcp.ConnectMode)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), cloudresourcesv1beta1.GcpNfsVolumeProcessing, nfsVol.Status.State)
}

func (s *modifyKcpNfsInstanceSuite) TestCreateNfsInstanceNoLocationNoZones() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj := gcpNfsVolume.DeepCopy()
	obj.Spec.Location = ""
	state, err := factory.newStateWith(obj)
	s.Nil(err)
	state.KcpIpRange = &kcpIpRange
	state.Scope = &kcpScope
	state.Scope.Spec.Scope.Gcp.Workers = nil
	//Invoke modifyKcpNfsInstance
	err, _ = modifyKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), composed.StopAndForget, err)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, nfsVol)

	//validate Status.ID of the GcpNfsVolume
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), nfsVol.Name, nfsVol.Status.Id)
	assert.Equal(s.T(), nfsVol.Status.Conditions[0].Type, cloudresourcesv1beta1.ConditionTypeError)
	assert.Equal(s.T(), nfsVol.Status.Conditions[0].Reason, cloudresourcesv1beta1.ConditionReasonNoWorkerZones)
}

func (s *modifyKcpNfsInstanceSuite) TestModifyNfsInstance() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	nfsVol := gcpNfsVolume.DeepCopy()
	state, err := factory.newStateWith(nfsVol)
	s.Nil(err)
	state.KcpIpRange = &kcpIpRange
	state.KcpNfsInstance = &gcpNfsInstance

	//Update GcpNfsVolume with new CapacityGb
	nfsVol.Spec.CapacityGb = 2048
	err = factory.skrCluster.K8sClient().Update(ctx, nfsVol)
	assert.Nil(s.T(), err)

	//Invoke modifyKcpNfsInstance
	err, _ = modifyKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), err, composed.StopWithRequeue)

	//Get the KcpNfsInstance using theGcpNfsVolume.Status.Id
	nfsInstance := cloudcontrolv1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: nfsVol.Status.Id, Namespace: kymaRef.Namespace}, &nfsInstance)
	assert.Nil(s.T(), err)

	//Validate KCPNfsInstance attributes.
	assert.Equal(s.T(), nfsVol.Spec.CapacityGb, nfsInstance.Spec.Instance.Gcp.CapacityGb)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), cloudresourcesv1beta1.GcpNfsVolumeProcessing, nfsVol.Status.State)
}

func (s *modifyKcpNfsInstanceSuite) TestModifyNfsInstanceNoLocation() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	nfsVol := gcpNfsVolume.DeepCopy()
	nfsVol.Spec.Location = ""
	state, err := factory.newStateWith(nfsVol)
	s.Nil(err)
	state.KcpIpRange = &kcpIpRange
	state.KcpNfsInstance = &gcpNfsInstance
	state.Scope = &kcpScope

	//Update GcpNfsVolume with new CapacityGb
	nfsVol.Spec.CapacityGb = 2048
	err = factory.skrCluster.K8sClient().Update(ctx, nfsVol)
	assert.Nil(s.T(), err)

	//Invoke modifyKcpNfsInstance
	err, _ = modifyKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), err, composed.StopWithRequeue)

	//Get the KcpNfsInstance using theGcpNfsVolume.Status.Id
	nfsInstance := cloudcontrolv1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: nfsVol.Status.Id, Namespace: kymaRef.Namespace}, &nfsInstance)
	assert.Nil(s.T(), err)

	//Validate KCPNfsInstance attributes.
	assert.Equal(s.T(), nfsVol.Spec.CapacityGb, nfsInstance.Spec.Instance.Gcp.CapacityGb)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), cloudresourcesv1beta1.GcpNfsVolumeProcessing, nfsVol.Status.State)
}

func (s *modifyKcpNfsInstanceSuite) TestWhenNfsVolumeDeleting() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(&deletedGcpNfsVolume)
	s.Nil(err)

	err, _ctx := modifyKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)
}

func (s *modifyKcpNfsInstanceSuite) TestWhenNfsVolumeNotChanged() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newState()
	assert.Nil(s.T(), err)
	state.KcpNfsInstance = &gcpNfsInstance

	err, _ctx := modifyKcpNfsInstance(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)
}

func TestModifyKcpNfsInstance(t *testing.T) {
	suite.Run(t, new(modifyKcpNfsInstanceSuite))
}
