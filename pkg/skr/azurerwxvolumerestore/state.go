package azurerwxvolumerestore

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	composed.State
	kymaRef               klog.ObjectRef
	kcpCluster            composed.StateCluster
	storageClient         client.Client
	scope                 *cloudcontrolv1beta1.Scope
	azureRwxVolumeBackup  *cloudresourcesv1beta1.AzureRwxVolumeBackup
	resourceGroupName     string
	storageAccountName    string
	fileShareName         string
	pvc                   *v1.PersistentVolumeClaim
	storageClientProvider azureclient.ClientProvider[client.Client]
}

type stateFactory struct {
	baseStateFactory      composed.StateFactory
	kymaRef               klog.ObjectRef
	kcpCluster            composed.StateCluster
	storageClientProvider azureclient.ClientProvider[client.Client]
}

func (f *stateFactory) NewState(req ctrl.Request) *State {

	return &State{
		State:                 f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AzureRwxVolumeBackup{}),
		kymaRef:               f.kymaRef,
		kcpCluster:            f.kcpCluster,
		storageClientProvider: f.storageClientProvider,
	}
}

func newStateFactory(
	baseStateFactory composed.StateFactory,
	kymaRef klog.ObjectRef,
	kcpCluster composed.StateCluster,
	storageClientProvider azureclient.ClientProvider[client.Client],
) *stateFactory {
	return &stateFactory{
		baseStateFactory:      baseStateFactory,
		kymaRef:               kymaRef,
		kcpCluster:            kcpCluster,
		storageClientProvider: storageClientProvider,
	}
}

func (s *State) ObjAsAzureRwxVolumeRestore() *cloudresourcesv1beta1.AzureRwxVolumeRestore {
	return s.Obj().(*cloudresourcesv1beta1.AzureRwxVolumeRestore)
}
