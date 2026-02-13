package gcpnfsvolume

import (
	"context"
	"fmt"
	"net/http/httptest"
	"time"

	gcpnfsbackupclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v1"

	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var kymaRef = klog.ObjectRef{
	Name:      "skr",
	Namespace: "test",
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
		Id:         "test-gcp-nfs-instance",
		Hosts:      []string{"10.20.30.2"},
		CapacityGb: 1024,
		Capacity:   resource.MustParse("1024Gi"),
		Conditions: []metav1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS is instance is ready",
			},
		},
	},
}

var deletedGcpNfsVolume = cloudresourcesv1beta1.GcpNfsVolume{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "deleted-gcp-nfs-volume",
		Namespace: "test",
		DeletionTimestamp: &metav1.Time{
			Time: time.Now(),
		},
		Finalizers: []string{"test-finalizer"},
	},
	Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{},
	Status: cloudresourcesv1beta1.GcpNfsVolumeStatus{
		Id:         "to-delete-gcp-nfs-instance",
		Hosts:      []string{"10.20.30.4"},
		CapacityGb: 1024,
		Capacity:   resource.MustParse("1024Gi"),
	},
}

var pvGcpNfsVolume = corev1.PersistentVolume{
	ObjectMeta: metav1.ObjectMeta{
		Name: fmt.Sprintf("%s--%s", gcpNfsVolume.Namespace, gcpNfsVolume.Name),
		Labels: map[string]string{
			cloudresourcesv1beta1.LabelNfsVolName: gcpNfsVolume.Name,
			cloudresourcesv1beta1.LabelNfsVolNS:   gcpNfsVolume.Namespace,
		},
		Finalizers: []string{"kubernetes.io/pv-protection"},
	},
	Spec: corev1.PersistentVolumeSpec{
		Capacity: corev1.ResourceList{
			"storage": resource.Quantity{
				Format: "1024Gi",
			},
		},
		PersistentVolumeSource: corev1.PersistentVolumeSource{
			NFS: &corev1.NFSVolumeSource{
				Server: gcpNfsVolume.Status.Hosts[0],
				Path:   fmt.Sprintf("/%s", gcpNfsVolume.Spec.FileShareName),
			},
		},
	},
	Status: corev1.PersistentVolumeStatus{
		Phase: "Available",
	},
}

var pvDeletingGcpNfsVolume = corev1.PersistentVolume{
	ObjectMeta: metav1.ObjectMeta{
		Name: fmt.Sprintf("%s--%s", deletedGcpNfsVolume.Namespace, deletedGcpNfsVolume.Name),
		Labels: map[string]string{
			cloudresourcesv1beta1.LabelNfsVolName: deletedGcpNfsVolume.Name,
			cloudresourcesv1beta1.LabelNfsVolNS:   deletedGcpNfsVolume.Namespace,
		},
		Finalizers: []string{"kubernetes.io/pv-protection"},
	},
	Spec: corev1.PersistentVolumeSpec{
		Capacity: nil,
		PersistentVolumeSource: corev1.PersistentVolumeSource{
			NFS: &corev1.NFSVolumeSource{
				Server:   deletedGcpNfsVolume.Status.Hosts[0],
				Path:     fmt.Sprintf("/%s", gcpNfsVolume.Spec.FileShareName),
				ReadOnly: false,
			},
		},
	},
	Status: corev1.PersistentVolumeStatus{
		Phase: "Available",
	},
}

var kcpScope = cloudcontrolv1beta1.Scope{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: kymaRef.Namespace,
		Name:      kymaRef.Name,
	},
	Spec: cloudcontrolv1beta1.ScopeSpec{
		KymaName:  kymaRef.Name,
		ShootName: kymaRef.Namespace,
		Region:    "us-west1",
		Provider:  cloudcontrolv1beta1.ProviderGCP,
		Scope: cloudcontrolv1beta1.ScopeInfo{
			Gcp: &cloudcontrolv1beta1.GcpScope{
				Project:    "test-project",
				VpcNetwork: fmt.Sprintf("shoot--%s--%s", "test-project", kymaRef.Name),
				Network: cloudcontrolv1beta1.GcpNetwork{
					Nodes:    "10.250.0.0/22",
					Pods:     "10.96.0.0/13",
					Services: "10.104.0.0/13",
				},
				Workers: []cloudcontrolv1beta1.GcpWorkers{
					{
						Zones: []string{"us-central1-a", "us-central1-b", "us-central1-c"},
					},
				},
			},
		},
	},
}

var kcpIpRange = cloudcontrolv1beta1.IpRange{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-ip-range",
		Namespace: kymaRef.Namespace,
		Labels: map[string]string{
			cloudcontrolv1beta1.LabelKymaName:   kymaRef.Name,
			cloudcontrolv1beta1.LabelRemoteName: gcpNfsVolume.Spec.IpRange.Name,
		},
	},
	Spec: cloudcontrolv1beta1.IpRangeSpec{
		RemoteRef: cloudcontrolv1beta1.RemoteRef{
			Name: gcpNfsVolume.Spec.IpRange.Name,
		},
		Scope: cloudcontrolv1beta1.ScopeRef{
			Name: kymaRef.Name,
		},
		Cidr: "10.20.30.0/24",
		Options: cloudcontrolv1beta1.IpRangeOptions{
			Gcp: &cloudcontrolv1beta1.IpRangeGcp{
				Purpose: cloudcontrolv1beta1.GcpPurposePSA,
			},
		},
	},
}

var skrIpRange = cloudresourcesv1beta1.IpRange{
	ObjectMeta: metav1.ObjectMeta{
		Name: gcpNfsVolume.Spec.IpRange.Name,
	},
	Spec: cloudresourcesv1beta1.IpRangeSpec{
		Cidr: "10.20.30.0/24",
	},
}

var gcpNfsInstance = cloudcontrolv1beta1.NfsInstance{
	ObjectMeta: metav1.ObjectMeta{
		Name:      gcpNfsVolume.Status.Id,
		Namespace: kymaRef.Namespace,
		Labels: map[string]string{
			cloudcontrolv1beta1.LabelKymaName:        kymaRef.Name,
			cloudcontrolv1beta1.LabelRemoteName:      gcpNfsVolume.Name,
			cloudcontrolv1beta1.LabelRemoteNamespace: gcpNfsVolume.Namespace,
		},
	},
	Spec: cloudcontrolv1beta1.NfsInstanceSpec{
		RemoteRef: cloudcontrolv1beta1.RemoteRef{
			Namespace: gcpNfsVolume.Namespace,
			Name:      gcpNfsVolume.Name,
		},
		IpRange: cloudcontrolv1beta1.IpRangeRef{
			Name: gcpNfsVolume.Spec.IpRange.Name,
		},
		Scope: cloudcontrolv1beta1.ScopeRef{
			Name: kymaRef.Name,
		},
		Instance: cloudcontrolv1beta1.NfsInstanceInfo{
			Gcp: &cloudcontrolv1beta1.NfsInstanceGcp{
				Tier:          cloudcontrolv1beta1.GcpFileTier(gcpNfsVolume.Spec.Tier),
				FileShareName: gcpNfsVolume.Spec.FileShareName,
				CapacityGb:    gcpNfsVolume.Spec.CapacityGb,
				ConnectMode:   cloudcontrolv1beta1.PRIVATE_SERVICE_ACCESS,
			},
		},
	},
	Status: cloudcontrolv1beta1.NfsInstanceStatus{
		State: "Ready",
		Id:    "gcp-filestore-1",
		Conditions: []metav1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS is instance is ready",
			},
		},
		Hosts:      []string{"10.20.30.2"},
		CapacityGb: gcpNfsVolume.Spec.CapacityGb,
		Capacity:   resource.MustParse(fmt.Sprintf("%dGi", gcpNfsVolume.Spec.CapacityGb)),
	},
}

var gcpNfsInstanceToDelete = cloudcontrolv1beta1.NfsInstance{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "to-delete-gcp-nfs-instance",
		Namespace: kymaRef.Namespace,
		Labels: map[string]string{
			cloudcontrolv1beta1.LabelKymaName:        kymaRef.Name,
			cloudcontrolv1beta1.LabelRemoteName:      "deleted-gcp-nfs-volume",
			cloudcontrolv1beta1.LabelRemoteNamespace: "test",
		},
		Finalizers: []string{api.CommonFinalizerDeletionHook},
	},
	Spec: cloudcontrolv1beta1.NfsInstanceSpec{
		RemoteRef: cloudcontrolv1beta1.RemoteRef{
			Namespace: "test",
			Name:      "deleted-gcp-nfs-volume",
		},
		IpRange: cloudcontrolv1beta1.IpRangeRef{
			Name: "test-gcp-ip-range",
		},
		Scope: cloudcontrolv1beta1.ScopeRef{
			Name: kymaRef.Name,
		},
		Instance: cloudcontrolv1beta1.NfsInstanceInfo{
			Gcp: &cloudcontrolv1beta1.NfsInstanceGcp{
				Tier:          cloudcontrolv1beta1.GcpFileTier(gcpNfsVolume.Spec.Tier),
				FileShareName: gcpNfsVolume.Spec.FileShareName,
				CapacityGb:    gcpNfsVolume.Spec.CapacityGb,
				ConnectMode:   cloudcontrolv1beta1.PRIVATE_SERVICE_ACCESS,
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
		Source: cloudresourcesv1beta1.GcpNfsVolumeBackupSource{
			Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
				Name:      "test-gcp-nfs-volume",
				Namespace: "test",
			},
		},
	},
	Status: cloudresourcesv1beta1.GcpNfsVolumeBackupStatus{
		Location: "us-west1",
		Conditions: []metav1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS backup is ready",
			},
		},
		Id: "backup-uuid",
	},
}

type testStateFactory struct {
	factory             StateFactory
	skrCluster          composed.StateCluster
	kcpCluster          composed.StateCluster
	gcpNfsVolume        *cloudresourcesv1beta1.GcpNfsVolume
	deletedGcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume
}

func newTestStateFactoryWithObject(fakeHttpServer *httptest.Server, backup *cloudresourcesv1beta1.GcpNfsVolumeBackup, volumes ...*cloudresourcesv1beta1.GcpNfsVolume) (*testStateFactory, error) {

	kcpClient := fake.NewClientBuilder().
		WithScheme(commonscheme.KcpScheme).
		WithObjects(&gcpNfsInstance).
		WithStatusSubresource(&gcpNfsInstance).
		WithObjects(&gcpNfsInstanceToDelete).
		WithStatusSubresource(&gcpNfsInstanceToDelete).
		WithObjects(&kcpScope).
		WithStatusSubresource(&kcpScope).
		Build()
	kcpCluster := composed.NewStateCluster(kcpClient, kcpClient, nil, commonscheme.KcpScheme)

	clientBuilder := fake.NewClientBuilder().
		WithScheme(commonscheme.SkrScheme)
	for _, volume := range volumes {
		clientBuilder = clientBuilder.WithObjects(volume).
			WithStatusSubresource(volume)

	}
	if backup.Name != "" {
		clientBuilder = clientBuilder.WithObjects(backup).
			WithStatusSubresource(backup)
	}
	skrClient := clientBuilder.Build()

	skrCluster := composed.NewStateCluster(skrClient, skrClient, nil, commonscheme.SkrScheme)

	nfsRestoreClientBackup := NewFakeFileBackupClientProvider(fakeHttpServer)
	env := abstractions.NewMockedEnvironment(map[string]string{"GCP_SA_JSON_KEY_PATH": "test"})

	factory := NewStateFactory(kymaRef, kcpCluster, skrCluster, nfsRestoreClientBackup, env)

	return &testStateFactory{
		factory:             factory,
		skrCluster:          skrCluster,
		kcpCluster:          kcpCluster,
		gcpNfsVolume:        &gcpNfsVolume,
		deletedGcpNfsVolume: &deletedGcpNfsVolume,
	}, nil

}

func newTestStateFactory(fakeHttpServer *httptest.Server) (*testStateFactory, error) {
	var backup cloudresourcesv1beta1.GcpNfsVolumeBackup
	return newTestStateFactoryWithObject(fakeHttpServer, &backup, &gcpNfsVolume, &deletedGcpNfsVolume)
}

func (f *testStateFactory) newState() (*State, error) {
	return f.newStateWith(&gcpNfsVolume)
}

func (f *testStateFactory) newStateWith(nfsVolume *cloudresourcesv1beta1.GcpNfsVolume) (*State, error) {
	return f.factory.NewState(context.Background(), composed.NewStateFactory(f.skrCluster).NewState(
		types.NamespacedName{
			Name:      nfsVolume.Name,
			Namespace: nfsVolume.Namespace,
		}, nfsVolume))
}

// Fake client doesn't support type "apply" for patching so falling back on update for unit tests.
func (s *State) PatchObjStatus(ctx context.Context) error {
	return s.Cluster().K8sClient().Status().Update(ctx, s.Obj())
}

func NewFakeFileBackupClientProvider(fakeHttpServer *httptest.Server) client.ClientProvider[gcpnfsbackupclientv1.FileBackupClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, credentialsFile string) (gcpnfsbackupclientv1.FileBackupClient, error) {
			fsClient, err := file.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(fakeHttpServer.URL))
			if err != nil {
				return nil, err
			}
			return gcpnfsbackupclientv1.NewFileBackupClient(fsClient), nil
		},
	)
}
