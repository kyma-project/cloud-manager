package testinfra

import (
	"context"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ClusterType string

const (
	ClusterTypeKcp    = ClusterType("kcp")
	ClusterTypeSkr    = ClusterType("skr")
	ClusterTypeGarden = ClusterType("garden")
)

type Infra interface {
	InfraEnv
	InfraDSL

	KCP() ClusterInfo
	SKR() ClusterInfo
	Garden() ClusterInfo

	Stop() error
}

type ClusterInfo interface {
	ClusterEnv
	ClusterDSL

	Scheme() *runtime.Scheme
	Client() ctrlclient.Client
	Cfg() *rest.Config
}

type InfraEnv interface {
	KcpManager() manager.Manager
	Registry() skrruntime.SkrRegistry
	Looper() skrruntime.SkrLooper
	AwsMock() awsmock.Server
	SkrKymaRef() klog.ObjectRef

	StartKcpControllers(ctx context.Context)
	StartSkrControllers(ctx context.Context)
	Ctx() context.Context
	stopControllers()
}

type ClusterEnv interface {
	Namespace() string
	ObjKey(name string) types.NamespacedName
}

type InfraDSL interface {
	GivenKymaCRExists(name string) error
	GivenGardenShootAwsExists(name string) error
	GivenScopeAwsExists(name string) error
	WhenKymaModuleStateUpdates(kymaName string, state util.KymaModuleState) error
}

type ClusterDSL interface {
	GivenSecretExists(name string, data map[string][]byte) error
	GivenNamespaceExists(name string) error
}
