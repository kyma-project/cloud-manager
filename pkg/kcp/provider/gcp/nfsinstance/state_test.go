package nfsinstance

import (
	"context"
	"net/http/httptest"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
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

func getDeletedGcpNfsInstance() *cloudcontrolv1beta1.NfsInstance {
	return &cloudcontrolv1beta1.NfsInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deleted-gcp-nfs-instance",
			Namespace: "kcp-system",
			DeletionTimestamp: &metav1.Time{
				Time: time.Now(),
			},
			Finalizers: []string{"test-finalizer"},
		},
		Spec:   cloudcontrolv1beta1.NfsInstanceSpec{},
		Status: cloudcontrolv1beta1.NfsInstanceStatus{},
	}
}

func getGcpNfsInstance() *cloudcontrolv1beta1.NfsInstance {
	return &cloudcontrolv1beta1.NfsInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gcp-nfs-instance",
			Namespace: kymaRef.Namespace,
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        kymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      "test-gcp-nfs-volume",
				cloudcontrolv1beta1.LabelRemoteNamespace: "test",
			},
		},
		Spec: cloudcontrolv1beta1.NfsInstanceSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: "test",
				Name:      "test-gcp-nfs-volume",
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: "test-gcp-ip-range",
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: kymaRef.Name,
			},
			Instance: cloudcontrolv1beta1.NfsInstanceInfo{
				Gcp: &cloudcontrolv1beta1.NfsInstanceGcp{
					Location:      "us-west1",
					Tier:          "BASIC_HDD",
					FileShareName: "vol1",
					CapacityGb:    1024,
					ConnectMode:   cloudcontrolv1beta1.PRIVATE_SERVICE_ACCESS,
				},
			},
		},
		Status: cloudcontrolv1beta1.NfsInstanceStatus{
			State: "Ready",
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             "True",
					LastTransitionTime: metav1.Time{Time: time.Now()},
					Reason:             "Ready",
					Message:            "NFS is instance is ready",
				},
			},
			Hosts:      []string{"10.0.0.2"},
			CapacityGb: 1024,
		},
	}
}

func getGcpNfsInstanceWithoutStatus() *cloudcontrolv1beta1.NfsInstance {
	var gcpNfsInstance2 = getGcpNfsInstance().DeepCopy()
	gcpNfsInstance2.Name = "test-gcp-nfs-instance-2"
	gcpNfsInstance2.Labels[cloudcontrolv1beta1.LabelRemoteName] = "test-gcp-nfs-volume-2"
	gcpNfsInstance2.Spec.RemoteRef.Name = "test-gcp-nfs-volume-2"
	gcpNfsInstance2.Status = cloudcontrolv1beta1.NfsInstanceStatus{}
	return gcpNfsInstance2
}

type testStateFactory struct {
	factory                StateFactory
	kcpCluster             composed.StateCluster
	nfsInstance            *cloudcontrolv1beta1.NfsInstance
	nfsInstanceNoCondition *cloudcontrolv1beta1.NfsInstance
	deletedNfsInstance     *cloudcontrolv1beta1.NfsInstance
	fakeHttpServer         *httptest.Server
}

func newTestStateFactory(fakeHttpServer *httptest.Server, objects ...*cloudcontrolv1beta1.NfsInstance) (*testStateFactory, error) {
	kcpClientBuilder := fake.NewClientBuilder().
		WithScheme(commonscheme.KcpScheme)
	for _, obj := range objects {
		kcpClientBuilder = kcpClientBuilder.WithObjects(obj).WithStatusSubresource(obj)
	}
	kcpClient := kcpClientBuilder.
		Build()
	kcpCluster := composed.NewStateCluster(kcpClient, kcpClient, nil, commonscheme.KcpScheme)
	fakeFileStoreClientProvider := NewFakeFilestoreClientProvider(fakeHttpServer)
	factory := NewStateFactory(fakeFileStoreClientProvider, abstractions.NewMockedEnvironment(map[string]string{"GCP_SA_JSON_KEY_PATH": "test"}))

	return &testStateFactory{
		factory:                factory,
		kcpCluster:             kcpCluster,
		nfsInstance:            getGcpNfsInstance(),
		nfsInstanceNoCondition: getGcpNfsInstanceWithoutStatus(),
		deletedNfsInstance:     getDeletedGcpNfsInstance(),
		fakeHttpServer:         fakeHttpServer,
	}, nil

}

type TestState struct {
	*State
	FakeHttpServer *httptest.Server
}

func (f *testStateFactory) newStateWith(ctx context.Context, nfsInstance *cloudcontrolv1beta1.NfsInstance, opIdentifier string) (*TestState, error) {
	focalState := focal.NewStateFactory().NewState(composed.NewStateFactory(f.kcpCluster).NewState(
		types.NamespacedName{
			Name:      nfsInstance.Name,
			Namespace: nfsInstance.Namespace,
		}, nfsInstance))
	focalState.SetScope(&cloudcontrolv1beta1.Scope{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kymaRef.Name,
			Namespace: kymaRef.Namespace,
		},
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Region: "us-west1",
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Gcp: &cloudcontrolv1beta1.GcpScope{
					Project:    "test-project",
					VpcNetwork: "test-vpc",
				},
			},
		},
	})
	typesState := newTypesState(focalState)

	state, err := f.factory.NewState(ctx, typesState)
	if err != nil {
		return nil, err
	}
	if opIdentifier != "" {
		state.ObjAsNfsInstance().Status.OpIdentifier = opIdentifier
	}
	if state.ObjAsNfsInstance().Spec.IpRange.Name != "" {
		state.SetIpRange(&cloudcontrolv1beta1.IpRange{
			Spec: cloudcontrolv1beta1.IpRangeSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Name: state.ObjAsNfsInstance().Spec.IpRange.Name,
				},
			},
		})
	}
	return &TestState{
		State:          state,
		FakeHttpServer: f.fakeHttpServer,
	}, nil
}

func NewFakeFilestoreClientProvider(fakeHttpServer *httptest.Server) client.ClientProvider[gcpnfsinstanceclient.FilestoreClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, saJsonKeyPath string) (gcpnfsinstanceclient.FilestoreClient, error) {
			fsClient, err := file.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(fakeHttpServer.URL))
			if err != nil {
				return nil, err
			}
			return gcpnfsinstanceclient.NewFilestoreClient(fsClient), nil
		},
	)
}

type typesState struct {
	focal.State

	ipRange *cloudcontrolv1beta1.IpRange
}

func (s *typesState) ObjAsNfsInstance() *cloudcontrolv1beta1.NfsInstance {
	return s.Obj().(*cloudcontrolv1beta1.NfsInstance)
}

func (s *typesState) IpRange() *cloudcontrolv1beta1.IpRange {
	return s.ipRange
}

func (s *typesState) SetIpRange(r *cloudcontrolv1beta1.IpRange) {
	s.ipRange = r
}

func newTypesState(focalState focal.State) nfsinstancetypes.State {
	return &typesState{State: focalState}
}
