package azurerwxvolumebackup

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
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
	client             client.Client
	clientProvider     azureclient.ClientProvider[client.Client]
	resourceGroupName  string
	storageAccountName string
	fileShareName      string
	pvc                *corev1.PersistentVolumeClaim
	vaultName          string
	scope              *cloudcontrolv1beta1.Scope
}

func (s *State) ObjAsAzureRwxVolumeBackup() *cloudresourcesv1beta1.AzureRwxVolumeBackup {
	return s.Obj().(*cloudresourcesv1beta1.AzureRwxVolumeBackup)
}

type stateFactory struct {
	baseStateFactory        composed.StateFactory
	commonScopeStateFactory commonscope.StateFactory
	clientProvider          azureclient.ClientProvider[client.Client]
}

func (f *stateFactory) NewState(req ctrl.Request) *State {

	return &State{
		State: f.commonScopeStateFactory.NewState(
			f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AzureRwxVolumeRestore{}),
		),
		clientProvider: f.clientProvider,
	}
}

func newStateFactory(
	baseStateFactory composed.StateFactory,
	commonScopeStateFactory commonscope.StateFactory,
	clientProvider azureclient.ClientProvider[client.Client],
) *stateFactory {
	return &stateFactory{
		baseStateFactory:        baseStateFactory,
		commonScopeStateFactory: commonScopeStateFactory,
		clientProvider:          clientProvider,
	}
}
