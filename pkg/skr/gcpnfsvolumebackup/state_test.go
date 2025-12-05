package gcpnfsvolumebackup

import (
	"context"
	"net/http/httptest"
	"time"

	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var kymaRef = klog.ObjectRef{
	Name:      "skr",
	Namespace: "test",
}

var scope = cloudcontrolv1beta1.Scope{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "skr",
		Namespace: "test",
	},
	Spec: cloudcontrolv1beta1.ScopeSpec{
		Provider: "gcp",
		Scope: cloudcontrolv1beta1.ScopeInfo{
			Gcp: &cloudcontrolv1beta1.GcpScope{
				Project:    "test-project",
				VpcNetwork: "test-network",
			},
		},
	},
}

var gcpNfsVolume = cloudresourcesv1beta1.GcpNfsVolume{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-gcp-nfs-volume",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{
		IpRange: cloudresourcesv1beta1.IpRangeRef{
			Name: "test-gcp-ip-range",
		},
		Tier:          "BASIC_HDD",
		FileShareName: "vol1",
		CapacityGb:    1024,
	},
	Status: cloudresourcesv1beta1.GcpNfsVolumeStatus{
		Location:   "us-west1",
		Id:         "test-gcp-nfs-instance",
		Hosts:      []string{"10.20.30.2"},
		CapacityGb: 1024,
		Conditions: []metav1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS instance is ready",
			},
		},
	},
}
var gcpNfsVolumeBackup = cloudresourcesv1beta1.GcpNfsVolumeBackup{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-gcp-nfs-volume-backup",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.GcpNfsVolumeBackupSpec{
		Location: "us-west1",
		Source: cloudresourcesv1beta1.GcpNfsVolumeBackupSource{
			Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
				Name:      "test-gcp-nfs-volume",
				Namespace: "test",
			},
		},
	},
	Status: cloudresourcesv1beta1.GcpNfsVolumeBackupStatus{
		Location: "us-west1",
		State:    "Ready",
		Conditions: []metav1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS backup is ready",
			},
		},
		Id: "cffd6896-0127-48a1-8a64-e07f6ad5c912",
	},
}

var deletingGpNfsVolumeBackup = cloudresourcesv1beta1.GcpNfsVolumeBackup{
	ObjectMeta: metav1.ObjectMeta{
		Name:              "test-gcp-nfs-volume-restore",
		Namespace:         "test",
		DeletionTimestamp: &metav1.Time{Time: time.Now()},
		Finalizers:        []string{api.CommonFinalizerDeletionHook},
	},
	Spec: cloudresourcesv1beta1.GcpNfsVolumeBackupSpec{
		Location: "us-west1",
		Source: cloudresourcesv1beta1.GcpNfsVolumeBackupSource{
			Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
				Name:      "test-gcp-nfs-volume",
				Namespace: "test",
			},
		},
	},
	Status: cloudresourcesv1beta1.GcpNfsVolumeBackupStatus{
		Location: "us-west1",
		State:    "Ready",
		Conditions: []metav1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS backup is ready",
			},
		},
		Id: "cffd6896-0127-48a1-8a64-e07f6ad5c912",
	},
}

type testStateFactory struct {
	factory                  StateFactory
	skrCluster               composed.StateCluster
	kcpCluster               composed.StateCluster
	fileBackupClientProvider client.ClientProvider[gcpnfsbackupclient.FileBackupClient]
	env                      abstractions.Environment
	fakeHttpServer           *httptest.Server
}

func NewFakeFileBackupClientProvider(fakeHttpServer *httptest.Server) client.ClientProvider[gcpnfsbackupclient.FileBackupClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, credentialsFile string) (gcpnfsbackupclient.FileBackupClient, error) {
			fsClient, err := file.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(fakeHttpServer.URL))
			if err != nil {
				return nil, err
			}
			return gcpnfsbackupclient.NewFileBackupClient(fsClient), nil
		},
	)
}

func newTestStateFactoryWithObj(fakeHttpServer *httptest.Server, gcpNfsVolumeBackup *cloudresourcesv1beta1.GcpNfsVolumeBackup) (*testStateFactory, error) {
	kcpClient := fake.NewClientBuilder().
		WithScheme(commonscheme.KcpScheme).
		WithObjects(&scope).
		Build()
	kcpCluster := composed.NewStateCluster(kcpClient, kcpClient, nil, commonscheme.KcpScheme)

	skrClient := fake.NewClientBuilder().
		WithScheme(commonscheme.SkrScheme).
		WithObjects(&gcpNfsVolume).
		WithStatusSubresource(&gcpNfsVolume).
		WithObjects(gcpNfsVolumeBackup).
		WithStatusSubresource(gcpNfsVolumeBackup).
		Build()
	skrCluster := composed.NewStateCluster(skrClient, skrClient, nil, commonscheme.SkrScheme)
	nfsBackupClient := NewFakeFileBackupClientProvider(fakeHttpServer)
	env := abstractions.NewMockedEnvironment(map[string]string{"GCP_SA_JSON_KEY_PATH": "test"})
	factory := NewStateFactory(kymaRef, kcpCluster, skrCluster, nfsBackupClient, env)

	return &testStateFactory{
		factory:                  factory,
		skrCluster:               skrCluster,
		kcpCluster:               kcpCluster,
		fileBackupClientProvider: nfsBackupClient,
		env:                      env,
		fakeHttpServer:           fakeHttpServer,
	}, nil

}

func (f *testStateFactory) newStateWith(nfsBackup *cloudresourcesv1beta1.GcpNfsVolumeBackup) (*State, error) {
	return f.factory.NewState(context.Background(), composed.NewStateFactory(f.skrCluster).NewState(
		types.NamespacedName{
			Name:      nfsBackup.Name,
			Namespace: nfsBackup.Namespace,
		}, nfsBackup))
}

// Fake client doesn't support type "apply" for patching so falling back on update for unit tests.
func (s *State) PatchObjStatus(ctx context.Context) error {
	return s.Cluster().K8sClient().Status().Update(ctx, s.Obj())
}
