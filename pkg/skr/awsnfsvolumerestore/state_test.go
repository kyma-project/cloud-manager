package awsnfsvolumerestore

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	"k8s.io/apimachinery/pkg/types"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsnfsvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumerestore/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"time"
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
		Provider: "aws",
		Region:   awsnfsvolumebackupclient.MockAwsRegion,
		Scope: cloudcontrolv1beta1.ScopeInfo{
			Aws: &cloudcontrolv1beta1.AwsScope{
				AccountId:  client.MockAwsAccount,
				VpcNetwork: "test-network",
			},
		},
	},
}

var awsNfsVolume = cloudresourcesv1beta1.AwsNfsVolume{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-aws-nfs-volume",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.AwsNfsVolumeSpec{
		IpRange: cloudresourcesv1beta1.IpRangeRef{
			Name: "test-aws-ip-range",
		},
	},
	Status: cloudresourcesv1beta1.AwsNfsVolumeStatus{
		Id:     "test-aws-nfs-instance",
		Server: "10.20.30.2",
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
var awsNfsVolumeBackup = cloudresourcesv1beta1.AwsNfsVolumeBackup{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-aws-nfs-volume-backup",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.AwsNfsVolumeBackupSpec{
		Source: cloudresourcesv1beta1.AwsNfsVolumeBackupSource{
			Volume: cloudresourcesv1beta1.VolumeRef{
				Name:      "test-aws-nfs-volume",
				Namespace: "test",
			},
		},
	},
	Status: cloudresourcesv1beta1.AwsNfsVolumeBackupStatus{
		State: "Ready",
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

var awsNfsVolumeRestore = cloudresourcesv1beta1.AwsNfsVolumeRestore{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-aws-nfs-volume-restore",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.AwsNfsVolumeRestoreSpec{
		Source: cloudresourcesv1beta1.AwsNfsVolumeRestoreSource{
			Backup: cloudresourcesv1beta1.BackupRef{
				Name:      "test-aws-nfs-volume-backup",
				Namespace: "test",
			},
		},
	},
	Status: cloudresourcesv1beta1.AwsNfsVolumeRestoreStatus{
		State: "Ready",
		Conditions: []metav1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Restore is Ready",
			},
		},
		JobId:            "cffd6896-0127-48a1-8a64-e07f6ad5c912",
		IdempotencyToken: "test-idempotency-token",
	},
}
var deletingAwsNfsVolumeRestore = cloudresourcesv1beta1.AwsNfsVolumeRestore{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-aws-nfs-volume-restore",
		Namespace: "test",
		DeletionTimestamp: &metav1.Time{
			Time: time.Now(),
		},
		Finalizers: []string{api.CommonFinalizerDeletionHook},
	},
	Spec: cloudresourcesv1beta1.AwsNfsVolumeRestoreSpec{
		Source: cloudresourcesv1beta1.AwsNfsVolumeRestoreSource{
			Backup: cloudresourcesv1beta1.BackupRef{
				Name:      "test-aws-nfs-volume-backup",
				Namespace: "test",
			},
		},
	},
	Status: cloudresourcesv1beta1.AwsNfsVolumeRestoreStatus{
		State: "Ready",
		Conditions: []metav1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Restore is Ready",
			},
		},
		JobId:            "cffd6896-0127-48a1-8a64-e07f6ad5c912",
		IdempotencyToken: "test-idempotency-token",
	},
}

type testStateFactory struct {
	*stateFactory
	kcpCluster composed.StateCluster
	skrCluster composed.StateCluster
}

func newStateFactoryWithObj(awsNfsVolumeRestore *cloudresourcesv1beta1.AwsNfsVolumeRestore) (*testStateFactory, error) {

	kcpScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(kcpScheme))
	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(kcpScheme))

	kcpClient := fake.NewClientBuilder().
		WithScheme(kcpScheme).
		WithObjects(&scope).
		//WithObjects(&awsNfsInstance).
		Build()
	kcpCluster := composed.NewStateCluster(kcpClient, kcpClient, nil, kcpScheme)

	skrScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))

	skrClient := fake.NewClientBuilder().
		WithScheme(skrScheme).
		WithObjects(&awsNfsVolume).
		WithStatusSubresource(&awsNfsVolume).
		WithObjects(&awsNfsVolumeBackup).
		WithStatusSubresource(&awsNfsVolumeBackup).
		WithObjects(awsNfsVolumeRestore).
		WithStatusSubresource(awsNfsVolumeRestore).
		Build()
	skrCluster := composed.NewStateCluster(skrClient, skrClient, nil, skrScheme)

	env := abstractions.NewMockedEnvironment(map[string]string{"GCP_SA_JSON_KEY_PATH": "test"})
	factory := newStateFactory(
		composed.NewStateFactory(skrCluster),
		commonscope.NewStateFactory(kcpCluster, kymaRef),
		client.NewMockClient(), env,
	)
	return &testStateFactory{
		stateFactory: factory,
		kcpCluster:   kcpCluster,
		skrCluster:   skrCluster,
	}, nil
}

func (f *testStateFactory) newStateWith(obj *cloudresourcesv1beta1.AwsNfsVolumeRestore) (*State, error) {
	return &State{
		State: f.commonScopeStateFactory.NewState(
			f.composedStateFactory.NewState(types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, obj),
		),
		awsClientProvider: f.awsClientProvider,
		env:               f.env,
	}, nil
}

// Fake client doesn't support type "apply" for patching so falling back on update for unit tests.
func (s *State) PatchObjStatus(ctx context.Context) error {
	return s.Cluster().K8sClient().Status().Update(ctx, s.Obj())
}
