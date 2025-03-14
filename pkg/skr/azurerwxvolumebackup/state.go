package azurerwxvolumebackup

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	client2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	AuthSecret         *corev1.Secret
	client             client.Client
	resourceGroupName  string
	storageAccountName string
	fileShareName      string
	pvc                *v1.PersistentVolumeClaim
}

func (s *State) ObjAsAzureRwxVolumeBackup() *cloudresourcesv1beta1.AzureRwxVolumeBackup {
	return s.Obj().(*cloudresourcesv1beta1.AzureRwxVolumeBackup)
}

type stateFactory struct {
	baseStateFactory composed.StateFactory
	kymaRef          klog.ObjectRef
	kcpCluster       composed.StateCluster
	skrCluster       composed.StateCluster
	clientProvider   client2.ClientProvider[client.Client]
}

func (f *stateFactory) NewState(req ctrl.Request) (*State, error) {
	ctx := context.Background()
	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	subscriptionId := "" // TODO: confirm
	tenantId := ""       // TODO: confirm
	c, err := f.clientProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return nil, err
	}
	return &State{
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AzureRwxVolumeBackup{}),
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
		SkrCluster: f.skrCluster,
		client:     c,
	}, nil
}

func newStateFactory(
	baseStateFactory composed.StateFactory,
	kymaRef klog.ObjectRef,
	kcpCluster composed.StateCluster,
	skrCluster composed.StateCluster,
	clientProvider client2.ClientProvider[client.Client],
) *stateFactory {
	return &stateFactory{
		baseStateFactory: baseStateFactory,
		kymaRef:          kymaRef,
		kcpCluster:       kcpCluster,
		skrCluster:       skrCluster,
		clientProvider:   clientProvider,
	}
}
