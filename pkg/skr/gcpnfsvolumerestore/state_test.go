package gcpnfsvolumerestore

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	client2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/option"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"time"
)

var kymaRef = klog.ObjectRef{
	Name:      "skr",
	Namespace: "test",
}

var scope = cloudcontrolv1beta1.Scope{
	ObjectMeta: v1.ObjectMeta{
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
	ObjectMeta: v1.ObjectMeta{
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
		Conditions: []v1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: v1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS instance is ready",
			},
		},
	},
}
var gcpNfsVolumeBackup = cloudresourcesv1beta1.GcpNfsVolumeBackup{
	ObjectMeta: v1.ObjectMeta{
		Name:      "test-gcp-nfs-volume-backup",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.GcpNfsVolumeBackupSpec{
		Source: cloudresourcesv1beta1.GcpNfsVolumeBackupSource{
			Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
				Name:      "test-gcp-nfs-volume",
				Namespace: "test",
			},
		},
	},
	Status: cloudresourcesv1beta1.GcpNfsVolumeBackupStatus{
		Location: "us-west1",
		Conditions: []v1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: v1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS backup is ready",
			},
		},
		Id: "cffd6896-0127-48a1-8a64-e07f6ad5c912",
	},
}

var gcpNfsVolumeRestore = cloudresourcesv1beta1.GcpNfsVolumeRestore{
	ObjectMeta: v1.ObjectMeta{
		Name:      "test-gcp-nfs-volume-restore",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.GcpNfsVolumeRestoreSpec{
		Destination: cloudresourcesv1beta1.GcpNfsVolumeRestoreDestination{
			Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
				Name:      "test-gcp-nfs-volume",
				Namespace: "test",
			},
		},
		Source: cloudresourcesv1beta1.GcpNfsVolumeRestoreSource{
			Backup: cloudresourcesv1beta1.GcpNfsVolumeBackupRef{
				Name:      "test-gcp-nfs-volume-backup",
				Namespace: "test",
			},
		},
	},
}

var deletingGpNfsVolumeRestore = cloudresourcesv1beta1.GcpNfsVolumeRestore{
	ObjectMeta: v1.ObjectMeta{
		Name:              "test-gcp-nfs-volume-restore",
		Namespace:         "test",
		DeletionTimestamp: &v1.Time{Time: time.Now()},
		Finalizers:        []string{cloudresourcesv1beta1.Finalizer},
	},
	Spec: cloudresourcesv1beta1.GcpNfsVolumeRestoreSpec{
		Destination: cloudresourcesv1beta1.GcpNfsVolumeRestoreDestination{
			Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
				Name:      "test-gcp-nfs-volume",
				Namespace: "test",
			},
		},
		Source: cloudresourcesv1beta1.GcpNfsVolumeRestoreSource{
			Backup: cloudresourcesv1beta1.GcpNfsVolumeBackupRef{
				Name:      "test-gcp-nfs-volume-backup",
				Namespace: "test",
			},
		},
	},
}

type testStateFactory struct {
	factory                   StateFactory
	skrCluster                composed.StateCluster
	kcpCluster                composed.StateCluster
	fileRestoreClientProvider client.ClientProvider[client2.FileRestoreClient]
	env                       abstractions.Environment
	gcpConfig                 *client.GcpConfig
	fakeHttpServer            *httptest.Server
}

func NewFakeFileRestoreClientProvider(fakeHttpServer *httptest.Server) client.ClientProvider[client2.FileRestoreClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, saJsonKeyPath string) (client2.FileRestoreClient, error) {
			fsClient, err := file.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(fakeHttpServer.URL))
			if err != nil {
				return nil, err
			}
			return client2.NewFileRestoreClient(fsClient), nil
		},
	)
}

func newTestStateFactoryWithObj(fakeHttpServer *httptest.Server, gcpNfsVolumeRestore *cloudresourcesv1beta1.GcpNfsVolumeRestore) (*testStateFactory, error) {

	kcpScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(kcpScheme))
	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(kcpScheme))

	kcpClient := fake.NewClientBuilder().
		WithScheme(kcpScheme).
		WithObjects(&scope).
		Build()
	kcpCluster := composed.NewStateCluster(kcpClient, kcpClient, nil, kcpScheme)

	skrScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))

	skrClient := fake.NewClientBuilder().
		WithScheme(skrScheme).
		WithObjects(&gcpNfsVolume).
		WithStatusSubresource(&gcpNfsVolume).
		WithObjects(&gcpNfsVolumeBackup).
		WithStatusSubresource(&gcpNfsVolumeBackup).
		WithObjects(gcpNfsVolumeRestore).
		WithStatusSubresource(gcpNfsVolumeRestore).
		Build()
	skrCluster := composed.NewStateCluster(skrClient, skrClient, nil, skrScheme)
	nfsRestoreClient := NewFakeFileRestoreClientProvider(fakeHttpServer)
	env := abstractions.NewMockedEnvironment(map[string]string{"GCP_SA_JSON_KEY_PATH": "test"})
	factory := NewStateFactory(kymaRef, kcpCluster, skrCluster, nfsRestoreClient, env)

	return &testStateFactory{
		factory:                   factory,
		skrCluster:                skrCluster,
		kcpCluster:                kcpCluster,
		fileRestoreClientProvider: nfsRestoreClient,
		env:                       env,
		gcpConfig:                 client.GetGcpConfig(env),
		fakeHttpServer:            fakeHttpServer,
	}, nil

}

func (f *testStateFactory) newStateWith(nfsRestore *cloudresourcesv1beta1.GcpNfsVolumeRestore) (*State, error) {
	return f.factory.NewState(context.Background(), composed.NewStateFactory(f.skrCluster).NewState(
		types.NamespacedName{
			Name:      nfsRestore.Name,
			Namespace: nfsRestore.Namespace,
		}, nfsRestore))
}

// Fake client doesn't support type "apply" for patching so falling back on update for unit tests.
func (s *State) PatchObjStatus(ctx context.Context) error {
	return s.Cluster().K8sClient().Status().Update(ctx, s.Obj())
}
