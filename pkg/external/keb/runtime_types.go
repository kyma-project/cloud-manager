package keb

import (
	"fmt"

	gardener "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	registrycache "github.com/kyma-project/kim-snatch/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="STATE",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="SHOOT-NAME",type=string,JSONPath=`.metadata.labels.kyma-project\.io/shoot-name`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

const (
	Finalizer                              = "runtime-controller.infrastructure-manager.kyma-project.io/deletion-hook"
	AnnotationGardenerCloudDelConfirmation = "confirmation.gardener.cloud/deletion"
)

const (
	LabelKymaInstanceID      = "kyma-project.io/instance-id"
	LabelKymaRuntimeID       = "kyma-project.io/runtime-id"
	LabelKymaShootName       = "kyma-project.io/shoot-name"
	LabelKymaRegion          = "kyma-project.io/region"
	LabelKymaName            = "operator.kyma-project.io/kyma-name"
	LabelKymaBrokerPlanID    = "kyma-project.io/broker-plan-id"
	LabelKymaBrokerPlanName  = "kyma-project.io/broker-plan-name"
	LabelKymaGlobalAccountID = "kyma-project.io/global-account-id"
	LabelKymaSubaccountID    = "kyma-project.io/subaccount-id"
	LabelKymaManagedBy       = "operator.kyma-project.io/managed-by"
	LabelKymaInternal        = "operator.kyma-project.io/internal"
	LabelKymaPlatformRegion  = "kyma-project.io/platform-region"
)

const (
	RuntimeStateReady       = "Ready"
	RuntimeStateFailed      = "Failed"
	RuntimeStatePending     = "Pending"
	RuntimeStateTerminating = "Terminating"
)

type RuntimeConditionType string

const (
	ConditionTypeRuntimeProvisioned     RuntimeConditionType = "Provisioned"
	ConditionTypeRuntimeKubeconfigReady RuntimeConditionType = "KubeconfigReady"
	ConditionTypeOidcAndCMsConfigured   RuntimeConditionType = "OidcAndConfigMapConfigured"
	ConditionTypeRuntimeConfigured      RuntimeConditionType = "Configured"
	ConditionTypeRuntimeDeprovisioned   RuntimeConditionType = "Deprovisioned"
)

type RuntimeConditionReason string

const (
	ConditionReasonProcessing    = RuntimeConditionReason("Processing")
	ConditionReasonProcessingErr = RuntimeConditionReason("ProcessingErr")

	ConditionReasonInitialized            = RuntimeConditionReason("Initialised")
	ConditionReasonShootCreationPending   = RuntimeConditionReason("Pending")
	ConditionReasonShootCreationCompleted = RuntimeConditionReason("ShootCreationCompleted")

	ConditionReasonGardenerCRCreated      = RuntimeConditionReason("GardenerClusterCRCreated")
	ConditionReasonGardenerCRReady        = RuntimeConditionReason("GardenerClusterCRReady")
	ConditionReasonConfigurationCompleted = RuntimeConditionReason("ConfigurationCompleted")
	ConditionReasonConfigurationErr       = RuntimeConditionReason("ConfigurationError")

	ConditionReasonGardenerCRDeleted       = RuntimeConditionReason("GardenerClusterCRDeleted")
	ConditionReasonGardenerShootDeleted    = RuntimeConditionReason("GardenerShootDeleted")
	ConditionReasonStructuredConfigDeleted = RuntimeConditionReason("StructuredConfigDeleted")
	ConditionReasonConversionError         = RuntimeConditionReason("ConversionErr")
	ConditionReasonCreationError           = RuntimeConditionReason("CreationErr")
	ConditionReasonGardenerError           = RuntimeConditionReason("GardenerErr")
	ConditionReasonKubernetesAPIErr        = RuntimeConditionReason("KubernetesErr")

	ConditionReasonAuditLogError = RuntimeConditionReason("AuditLogErr")

	ConditionReasonAdministratorsConfigured       = RuntimeConditionReason("AdministratorsConfigured")
	ConditionReasonOidcAndCMsConfigured           = RuntimeConditionReason("OidcAndConfigMapsConfigured")
	ConditionReasonOidcError                      = RuntimeConditionReason("OidcConfigurationErr")
	ConditionReasonKymaSystemNSError              = RuntimeConditionReason("KymaSystemCreationErr")
	ConditionReasonSeedNotFound                   = RuntimeConditionReason("SeedNotFound")
	ConditionReasonRegistryCacheError             = RuntimeConditionReason("RegistryCacheConfigurationErr")
	ConditionReasonSeedClusterPreProcessingError  = RuntimeConditionReason("SeedClusterPreProcessingErr")
	ConditionReasonSeedClusterPostProcessingError = RuntimeConditionReason("SeedClusterPostProcessingErr")
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.shoot.provider.type"
//+kubebuilder:printcolumn:name="Region",type="string",JSONPath=".spec.shoot.region"
//+kubebuilder:printcolumn:name="STATE",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Runtime is the Schema for the runtimes API
type Runtime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RuntimeSpec   `json:"spec,omitempty"`
	Status RuntimeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RuntimeList contains a list of Runtime
type RuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Runtime `json:"items"`
}

// RuntimeSpec defines the desired state of Runtime
type RuntimeSpec struct {
	Shoot    RuntimeShoot         `json:"shoot"`
	Security Security             `json:"security"`
	Caching  []ImageRegistryCache `json:"imageRegistryCache,omitempty"`
}

type ImageRegistryCache struct {
	Name      string                                `json:"name"`
	Namespace string                                `json:"namespace"`
	UID       string                                `json:"uid"`
	Config    registrycache.RegistryCacheConfigSpec `json:"config"`
}

// RuntimeStatus defines the observed state of Runtime
type RuntimeStatus struct {
	// State signifies current state of Runtime
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Pending;Ready;Terminating;Failed
	State State `json:"state,omitempty"`

	// List of status conditions to indicate the status of a ServiceInstance.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ProvisioningCompleted indicates if the initial provisioning of the cluster is completed
	ProvisioningCompleted bool `json:"provisioningCompleted,omitempty"`
}

type RuntimeShoot struct {
	Name                string                 `json:"name"`
	Purpose             gardener.ShootPurpose  `json:"purpose"`
	PlatformRegion      string                 `json:"platformRegion"`
	Region              string                 `json:"region"`
	LicenceType         *string                `json:"licenceType,omitempty"`
	SecretBindingName   string                 `json:"secretBindingName"`
	EnforceSeedLocation *bool                  `json:"enforceSeedLocation,omitempty"`
	Kubernetes          Kubernetes             `json:"kubernetes,omitempty"`
	Provider            Provider               `json:"provider"`
	Networking          Networking             `json:"networking"`
	ControlPlane        *gardener.ControlPlane `json:"controlPlane,omitempty"`
}

type Kubernetes struct {
	Version       *string   `json:"version,omitempty"`
	KubeAPIServer APIServer `json:"kubeAPIServer,omitempty"`
}

// OIDCConfig contains configuration settings for the OIDC provider.
// Note: Descriptions were taken from the Kubernetes documentation.
type OIDCConfig struct {
	gardener.OIDCConfig `json:",omitempty"`
	JWKS                []byte `json:"jwks,omitempty"`
}

type APIServer struct {
	OidcConfig           gardener.OIDCConfig `json:"oidcConfig,omitempty"`
	AdditionalOidcConfig *[]OIDCConfig       `json:"additionalOidcConfig,omitempty"`
}

type Provider struct {
	//+kubebuilder:validation:Enum=aws;azure;gcp;openstack
	Type                 string                `json:"type"`
	Workers              []gardener.Worker     `json:"workers"`
	AdditionalWorkers    *[]gardener.Worker    `json:"additionalWorkers,omitempty"`
	ControlPlaneConfig   *runtime.RawExtension `json:"controlPlaneConfig,omitempty"`
	InfrastructureConfig *runtime.RawExtension `json:"infrastructureConfig,omitempty"`
}

type Networking struct {
	Type     *string `json:"type,omitempty"`
	Pods     string  `json:"pods"`
	Nodes    string  `json:"nodes"`
	Services string  `json:"services"`
}

type Security struct {
	Administrators []string           `json:"administrators"`
	Networking     NetworkingSecurity `json:"networking"`
}

type NetworkingSecurity struct {
	Filter Filter `json:"filter"`
}

type Filter struct {
	Ingress *Ingress `json:"ingress,omitempty"`
	Egress  Egress   `json:"egress"`
}

// Ingress filtering can be enabled for `shoot-networking-fitler` extension with
// the blackholing feature, see https://github.com/gardener/gardener-extension-shoot-networking-filter/blob/master/docs/usage/shoot-networking-filter.md#ingress-filtering
type Ingress struct {
	// It means that the blackholing filtering is enabled on the per shoot level.
	Enabled bool `json:"enabled"`
}

// Egress filtering is a default filtering mode for `shoot-networking-fitler` extension.
type Egress struct {
	Enabled bool `json:"enabled"`
}

func init() {
	SchemeBuilder.Register(&Runtime{}, &RuntimeList{})
}

func (k *Runtime) UpdateStateReady(c RuntimeConditionType, r RuntimeConditionReason, msg string) {
	k.Status.State = RuntimeStateReady
	condition := metav1.Condition{
		Type:               string(c),
		Status:             "True",
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            msg,
	}
	meta.SetStatusCondition(&k.Status.Conditions, condition)
}

func (k *Runtime) UpdateStateDeletion(c RuntimeConditionType, r RuntimeConditionReason, status, msg string) {
	if status != "False" {
		k.Status.State = RuntimeStateTerminating
	} else {
		k.Status.State = RuntimeStateFailed
	}

	condition := metav1.Condition{
		Type:               string(c),
		Status:             metav1.ConditionStatus(status),
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            msg,
	}
	meta.SetStatusCondition(&k.Status.Conditions, condition)
}

func (k *Runtime) UpdateStatePending(c RuntimeConditionType, r RuntimeConditionReason, status, msg string) {
	if status == "False" {
		k.Status.State = RuntimeStateFailed
	} else {
		k.Status.State = RuntimeStatePending
	}

	condition := metav1.Condition{
		Type:               string(c),
		Status:             metav1.ConditionStatus(status),
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            msg,
	}
	meta.SetStatusCondition(&k.Status.Conditions, condition)
}

func (k *Runtime) UpdateStateProvisioningCompleted() {
	k.Status.ProvisioningCompleted = true
}

func (k *Runtime) IsProvisioningCompletedStatusSet() bool {
	return k.Status.ProvisioningCompleted
}

func (k *Runtime) IsStateWithConditionSet(runtimeState State, c RuntimeConditionType, r RuntimeConditionReason) bool {
	if k.Status.State != runtimeState {
		return false
	}

	return k.IsConditionSet(c, r)
}

func (k *Runtime) IsConditionSet(c RuntimeConditionType, r RuntimeConditionReason) bool {
	condition := meta.FindStatusCondition(k.Status.Conditions, string(c))
	if condition != nil && condition.Reason == string(r) {
		return true
	}
	return false
}

func (k *Runtime) IsStateWithConditionAndStatusSet(runtimeState State, c RuntimeConditionType, r RuntimeConditionReason, s metav1.ConditionStatus) bool {
	if k.Status.State != runtimeState {
		return false
	}

	return k.IsConditionSetWithStatus(c, r, s)
}

func (k *Runtime) IsConditionSetWithStatus(c RuntimeConditionType, r RuntimeConditionReason, s metav1.ConditionStatus) bool {
	condition := meta.FindStatusCondition(k.Status.Conditions, string(c))
	if condition != nil && condition.Reason == string(r) && condition.Status == s {
		return true
	}
	return false
}

func (k *Runtime) ValidateRequiredLabels() error {
	var requiredLabelKeys = []string{
		LabelKymaInstanceID,
		LabelKymaRuntimeID,
		LabelKymaRegion,
		LabelKymaName,
		LabelKymaBrokerPlanID,
		LabelKymaBrokerPlanName,
		LabelKymaGlobalAccountID,
		LabelKymaSubaccountID,
	}

	for _, key := range requiredLabelKeys {
		if k.Labels[key] == "" {
			return fmt.Errorf("missing required label %s", key)
		}
	}
	return nil
}
