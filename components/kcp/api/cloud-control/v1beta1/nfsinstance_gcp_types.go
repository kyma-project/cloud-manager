package v1beta1

type NfsOptionsGcp struct {
	// +kubebuilder:validation:Required
	Location string `json:"location"`

	// +kubebuilder:default=BASIC_HDD
	Tier GcpFileTier `json:"tier"`

	// +kubebuilder:validation:Pattern="^[a-z][a-z0-9_]*[a-z0-9]$"
	// +kubebuilder:default=vol1
	FileShareName string `json:"fileShareName"`

	// +kubebuilder:default=1024
	CapacityGb int `json:"capacityGb"`

	// +kubebuilder:default=PRIVATE_SERVICE_ACCESS
	ConnectMode GcpConnectMode `json:"connectMode"`
}

// +kubebuilder:validation:Enum=BASIC_HDD;BASIC_SSD;HIGH_SCALE_SSD;ENTERPRISE;ZONAL;REGIONAL
type GcpFileTier string

const (
	BASIC_HDD      = GcpFileTier("BASIC_HDD")
	BASIC_SSD      = GcpFileTier("BASIC_SSD")
	HIGH_SCALE_SSD = GcpFileTier("HIGH_SCALE_SSD")
	ENTERPRISE     = GcpFileTier("ENTERPRISE")
	ZONAL          = GcpFileTier("ZONAL")
	REGIONAL       = GcpFileTier("REGIONAL")
)

// +kubebuilder:validation:Enum=DIRECT_PEERING;PRIVATE_SERVICE_ACCESS
type GcpConnectMode string

const (
	DIRECT_PEERING         = GcpConnectMode("DIRECT_PEERING")
	PRIVATE_SERVICE_ACCESS = GcpConnectMode("PRIVATE_SERVICE_ACCESS")
)
