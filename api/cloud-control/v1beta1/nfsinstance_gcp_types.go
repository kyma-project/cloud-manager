package v1beta1

type NfsOptionsGcp struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Location is immutable."
	Location string `json:"location"`

	// +kubebuilder:default=BASIC_HDD
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Tier is immutable."
	Tier GcpFileTier `json:"tier"`

	// +kubebuilder:validation:Pattern="^[a-z][a-z0-9_]*[a-z0-9]$"
	// +kubebuilder:default=vol1
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="FileShareName is immutable."
	FileShareName string `json:"fileShareName"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="SourceBackup is immutable."
	SourceBackup string `json:"sourceBackup,omitempty"`

	// +kubebuilder:default=1024
	CapacityGb int `json:"capacityGb"`

	// +kubebuilder:default=PRIVATE_SERVICE_ACCESS
	ConnectMode GcpConnectMode `json:"connectMode"`
}

// +kubebuilder:validation:Enum=BASIC_HDD;BASIC_SSD;ZONAL;REGIONAL
type GcpFileTier string

const (
	BASIC_HDD = GcpFileTier("BASIC_HDD")
	BASIC_SSD = GcpFileTier("BASIC_SSD")
	ZONAL     = GcpFileTier("ZONAL")
	REGIONAL  = GcpFileTier("REGIONAL")
)

// +kubebuilder:validation:Enum=DIRECT_PEERING;PRIVATE_SERVICE_ACCESS
type GcpConnectMode string

const (
	DIRECT_PEERING         = GcpConnectMode("DIRECT_PEERING")
	PRIVATE_SERVICE_ACCESS = GcpConnectMode("PRIVATE_SERVICE_ACCESS")
)
