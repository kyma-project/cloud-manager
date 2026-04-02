package azurerwxvolumerestore

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	corev1 "k8s.io/api/core/v1"
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

func (f *stateFactory) NewState(ctx context.Context, req ctrl.Request) (*State, error) {
	scopeState, err := f.commonScopeStateFactory.NewState(
		ctx,
		req.NamespacedName,
		f.composedStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AzureRwxVolumeRestore{}),
	)
	if err != nil {
		return nil, err
	}
	return &State{
		State:                 scopeState,
		storageClientProvider: f.storageClientProvider,
	}, nil
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
