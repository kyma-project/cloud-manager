package awsnfsvolumebackup

import (
	commonScope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	"k8s.io/apimachinery/pkg/types"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	ObjectMeta: v1.ObjectMeta{
		Name:      "skr",
		Namespace: "test",
	},
	Spec: cloudcontrolv1beta1.ScopeSpec{
		Provider: "aws",
		Scope: cloudcontrolv1beta1.ScopeInfo{
			Aws: &cloudcontrolv1beta1.AwsScope{
				AccountId:  "test-project",
				VpcNetwork: "test-network",
			},
		},
	},
}

var awsNfsVolume = cloudresourcesv1beta1.AwsNfsVolume{
	ObjectMeta: v1.ObjectMeta{
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
var awsNfsVolumeBackup = cloudresourcesv1beta1.AwsNfsVolumeBackup{
	ObjectMeta: v1.ObjectMeta{
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

var deletingGpNfsVolumeBackup = cloudresourcesv1beta1.AwsNfsVolumeBackup{
	ObjectMeta: v1.ObjectMeta{
		Name:              "test-aws-nfs-volume-restore",
		Namespace:         "test",
		DeletionTimestamp: &v1.Time{Time: time.Now()},
		Finalizers:        []string{cloudresourcesv1beta1.Finalizer},
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

type testStateFactory struct {
	*stateFactory
}

func newStateFactoryWithObj(awsNfsVolumeBackup *cloudresourcesv1beta1.AwsNfsVolumeBackup) (*testStateFactory, error) {

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
		WithObjects(&awsNfsVolume).
		WithStatusSubresource(&awsNfsVolume).
		WithObjects(awsNfsVolumeBackup).
		WithStatusSubresource(awsNfsVolumeBackup).
		Build()
	skrCluster := composed.NewStateCluster(skrClient, skrClient, nil, skrScheme)

	env := abstractions.NewMockedEnvironment(map[string]string{"GCP_SA_JSON_KEY_PATH": "test"})
	factory := newStateFactory(
		composed.NewStateFactory(skrCluster),
		commonScope.NewStateFactory(kcpCluster, kymaRef),
		client.NewMockClient(), env,
	)
	return &testStateFactory{stateFactory: factory}, nil
}

func (f *testStateFactory) newStateWith(obj *cloudresourcesv1beta1.AwsNfsVolumeBackup) (*State, error) {
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
