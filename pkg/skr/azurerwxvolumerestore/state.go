package azurerwxvolumerestore

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	commonscope.State
	storageClient         client.Client
	azureRwxVolumeBackup  *cloudresourcesv1beta1.AzureRwxVolumeBackup
	resourceGroupName     string
	storageAccountName    string
	fileShareName         string
	pvc                   *corev1.PersistentVolumeClaim
	storageClientProvider azureclient.ClientProvider[client.Client]
}

type stateFactory struct {
	composedStateFactory    composed.StateFactory
	commonScopeStateFactory commonscope.StateFactory
	storageClientProvider   azureclient.ClientProvider[client.Client]
}

func (f *stateFactory) NewState(req ctrl.Request) *State {

	return &State{
		State: f.commonScopeStateFactory.NewState(
			f.composedStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AzureRwxVolumeRestore{}),
		),
		storageClientProvider: f.storageClientProvider,
	}
}

func newStateFactory(
	baseStateFactory composed.StateFactory,
	commonScopeStateFactory commonscope.StateFactory,
	storageClientProvider azureclient.ClientProvider[client.Client],
) *stateFactory {
	return &stateFactory{
		composedStateFactory:    baseStateFactory,
		commonScopeStateFactory: commonScopeStateFactory,
		storageClientProvider:   storageClientProvider,
	}
}

func (s *State) ObjAsAzureRwxVolumeRestore() *cloudresourcesv1beta1.AzureRwxVolumeRestore {
	return s.Obj().(*cloudresourcesv1beta1.AzureRwxVolumeRestore)
}

var kymaRef = klog.ObjectRef{
	Name:      "skr",
	Namespace: "test",
}
