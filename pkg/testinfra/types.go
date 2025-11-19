package testinfra

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/config"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	azuremock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/mock"
	gcpmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/mock"
	sapmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/mock"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Infra interface {
	InfraEnv
	InfraDSL

	ProjectRootDir() string

	KCP() ClusterInfo
	SKR() ClusterInfo
	Garden() ClusterInfo

	Stop() error

	FinalizerReport()
}

type ClusterInfo interface {
	ClusterEnv
	ClusterDSL

	Scheme() *runtime.Scheme
	Client() ctrlclient.Client
	Cfg() *rest.Config
	Kubeconfig() []byte
	KubeconfigFilePath() string
	EnsureCrds(ctx context.Context) error
}

type InfraEnv interface {
	KcpManager() manager.Manager
	Registry() skrruntime.SkrRegistry
	ActiveSkrCollection() skrruntime.ActiveSkrCollection
	AwsMock() awsmock.Server
	GcpMock() gcpmock.Server
	AzureMock() azuremock.Server
	SapMock() sapmock.Server
	SkrKymaRef() klog.ObjectRef
	SkrRunner() skrruntime.SkrRunner
	Config() config.Config

	StartKcpControllers(ctx context.Context)
	KcpWaitForCacheSync(ctx context.Context) error
	StartSkrControllers(ctx context.Context)
	Ctx() context.Context
	stopControllers()
}

type ClusterEnv interface {
	Namespace() string
	ObjKey(name string) types.NamespacedName
}

type InfraDSL interface {
	GivenSkrIpRangeExists(ctx context.Context, ns, name, cidr string, id string, conditions ...metav1.Condition) error
	WhenSkrIpRangeIsCreated(ctx context.Context, ns, name, cidr string, id string, conditions ...metav1.Condition) error

	WhenKymaModuleStateUpdates(kymaName string, state util.KymaModuleState) error
	GivenGardenShootGcpExists(name string) error
	GivenScopeGcpExists(name string) error
}

type ClusterDSL interface {
	GivenSecretExists(name string, data map[string][]byte) error
	GivenNamespaceExists(name string) error

	GivenConditionIsSet(ctx context.Context, obj composed.ObjWithConditions, conditions ...metav1.Condition) error
	WhenConditionIsSet(ctx context.Context, obj composed.ObjWithConditions, conditions ...metav1.Condition) error
}
