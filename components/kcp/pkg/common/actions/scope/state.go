package scope

import (
	"errors"
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerClient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/scope/client"
	awsClient "github.com/kyma-project/cloud-resources/components/kcp/pkg/provider/aws/client"
	kubernetesClient "k8s.io/client-go/kubernetes"
)

var canNotSetProviderWhenScopeIsSet error

func init() {
	canNotSetProviderWhenScopeIsSet = errors.New("can not set provider when scope is set")
}

type State interface {
	focal.State
	FileReader() abstractions.FileReader
	AwsStsClientProvider() awsClient.GardenClientProvider[client.AwsStsClient]

	ShootName() string
	SetShootName(string)
	ShootNamespace() string
	SetShootNamespace(string)
	GardenerClient() gardenerClient.CoreV1beta1Interface
	SetGardenerClient(gardenerClient.CoreV1beta1Interface)
	GardenK8sClient() kubernetesClient.Interface
	SetGardenK8sClient(kubernetesClient.Interface)
	Provider() cloudresourcesv1beta1.ProviderType
	SetProvider(cloudresourcesv1beta1.ProviderType) error
	Shoot() *gardenerTypes.Shoot
	SetShoot(*gardenerTypes.Shoot)
	CredentialData() map[string]string
}

type StateFactory interface {
	CreateState(focalState focal.State) State
}

func NewStateFactory(fileReader abstractions.FileReader, awsStsClientProvider awsClient.GardenClientProvider[client.AwsStsClient]) StateFactory {
	return &stateFactory{
		fileReader:           fileReader,
		awsStsClientProvider: awsStsClientProvider,
	}
}

type stateFactory struct {
	fileReader           abstractions.FileReader
	awsStsClientProvider awsClient.GardenClientProvider[client.AwsStsClient]
}

func (f *stateFactory) CreateState(focalState focal.State) State {
	return newState(focalState, f.fileReader, f.awsStsClientProvider)
}

func newState(focalState focal.State, fileReader abstractions.FileReader, awsStsClientProvider awsClient.GardenClientProvider[client.AwsStsClient]) State {
	return &state{
		State:                focalState,
		fileReader:           fileReader,
		awsStsClientProvider: awsStsClientProvider,
		credentialData:       map[string]string{},
	}
}

type state struct {
	focal.State

	fileReader abstractions.FileReader

	shootName      string
	shootNamespace string

	gardenerClient  gardenerClient.CoreV1beta1Interface
	gardenK8sClient kubernetesClient.Interface

	provider       cloudresourcesv1beta1.ProviderType
	shoot          *gardenerTypes.Shoot
	credentialData map[string]string

	awsStsClientProvider awsClient.GardenClientProvider[client.AwsStsClient]
}

func (s *state) FileReader() abstractions.FileReader {
	return s.fileReader
}

func (s *state) AwsStsClientProvider() awsClient.GardenClientProvider[client.AwsStsClient] {
	return s.awsStsClientProvider
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

func (s *state) Provider() cloudresourcesv1beta1.ProviderType {
	return s.provider
}

func (s *state) SetProvider(providerType cloudresourcesv1beta1.ProviderType) error {
	if s.Scope() != nil {
		return canNotSetProviderWhenScopeIsSet
	}
	s.provider = providerType
	return nil
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
