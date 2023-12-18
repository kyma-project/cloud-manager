package scope

import (
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerClient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
	kubernetesClient "k8s.io/client-go/kubernetes"
)

type State interface {
	focal.State
	FileReader() abstractions.FileReader
	ShootName() string
	SetShootName(string)
	ShootNamespace() string
	SetShootNamespace(string)
	GardenerClient() gardenerClient.CoreV1beta1Interface
	SetGardenerClient(gardenerClient.CoreV1beta1Interface)
	GardenK8sClient() kubernetesClient.Interface
	SetGardenK8sClient(kubernetesClient.Interface)
	Provider() ProviderType
	SetProvider(ProviderType)
	Shoot() *gardenerTypes.Shoot
	SetShoot(*gardenerTypes.Shoot)
	CredentialData() map[string]string
}

func NewState(focalState focal.State, fileReader abstractions.FileReader) State {
	return &state{
		State:          focalState,
		fileReader:     fileReader,
		credentialData: map[string]string{},
	}
}

type state struct {
	focal.State

	fileReader abstractions.FileReader

	shootName      string
	shootNamespace string

	gardenerClient  gardenerClient.CoreV1beta1Interface
	gardenK8sClient kubernetesClient.Interface

	provider       ProviderType
	shoot          *gardenerTypes.Shoot
	credentialData map[string]string
}

func (s *state) FileReader() abstractions.FileReader {
	return s.fileReader
}

func (s *state) ShootName() string {
	return s.shootName
}

func (s *state) SetShootName(shootName string) {
	s.shootName = shootName
}

func (s *state) ShootNamespace() string {
	return s.shootNamespace
}

func (s *state) SetShootNamespace(shootNamespace string) {
	s.shootNamespace = shootNamespace
}

func (s *state) GardenerClient() gardenerClient.CoreV1beta1Interface {
	return s.gardenerClient
}

func (s *state) SetGardenerClient(client gardenerClient.CoreV1beta1Interface) {
	s.gardenerClient = client
}

func (s *state) GardenK8sClient() kubernetesClient.Interface {
	return s.gardenK8sClient
}

func (s *state) SetGardenK8sClient(client kubernetesClient.Interface) {
	s.gardenK8sClient = client
}

func (s *state) Provider() ProviderType {
	return s.provider
}

func (s *state) SetProvider(providerType ProviderType) {
	s.provider = providerType
}

func (s *state) Shoot() *gardenerTypes.Shoot {
	return s.shoot
}

func (s *state) SetShoot(shoot *gardenerTypes.Shoot) {
	s.shoot = shoot
}

func (s *state) CredentialData() map[string]string {
	return s.credentialData
}
