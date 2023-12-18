package scope

import (
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerClient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetesClient "k8s.io/client-go/kubernetes"
)

func NewState(focalState *focal.State, fileReader abstractions.FileReader) *State {
	return &State{
		State:      focalState,
		FileReader: fileReader,
	}
}

type State struct {
	*focal.State

	FileReader abstractions.FileReader

	KymaObj runtime.Unstructured

	ShootName      string
	ShootNamespace string

	GardenerClient  gardenerClient.CoreV1beta1Interface
	GardenK8sClient kubernetesClient.Interface

	Provider       ProviderType
	Shoot          *gardenerTypes.Shoot
	CredentialData map[string]string
}
