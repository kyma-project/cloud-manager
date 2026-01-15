package v1beta1

// NfsOptionsGcp defines GCP-specific NFS instance options
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_HDD" || (self.capacityGb >= 1024 && self.capacityGb <= 65433)), message="BASIC_HDD tier capacityGb must be between 1024 and 65433"
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_HDD" || size(self.fileShareName) <= 16), message="BASIC_HDD tier fileShareName length must be 16 or less characters"
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_HDD" || self.capacityGb >= oldSelf.capacityGb), message="BASIC_HDD tier capacityGb cannot be reduced"
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_SSD" || (self.capacityGb >= 2560 && self.capacityGb <= 65433)), message="BASIC_SSD tier capacityGb must be between 2560 and 65433"
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_SSD" || size(self.fileShareName) <= 16), message="BASIC_SSD tier fileShareName length must be 16 or less characters"
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_SSD" || self.capacityGb >= oldSelf.capacityGb), message="BASIC_SSD tier capacityGb cannot be reduced"
// +kubebuilder:validation:XValidation:rule=(self.tier != "ZONAL" || (self.capacityGb >= 1024 && self.capacityGb <= 9984 && (self.capacityGb - 1024) % 256 == 0 || self.capacityGb >= 10240 && self.capacityGb <= 102400 && (self.capacityGb - 10240) % 2560 == 0)), message="ZONAL tier capacityGb must be between 1024 and 9984, and divisible by 256, or between 10240 and 102400, and divisible by 2560"
// +kubebuilder:validation:XValidation:rule=(self.tier != "ZONAL" || size(self.fileShareName) <= 32), message="ZONAL tier fileShareName length must be 32 or less characters"
// +kubebuilder:validation:XValidation:rule=(self.tier != "REGIONAL" || (self.capacityGb >= 1024 && self.capacityGb <= 9984 && (self.capacityGb - 1024) % 256 == 0 || self.capacityGb >= 10240 && self.capacityGb <= 102400 && (self.capacityGb - 10240) % 2560 == 0)), message="REGIONAL tier capacityGb must be between 1024 and 9984, and divisible by 256, or between 10240 and 102400, and divisible by 2560"
// +kubebuilder:validation:XValidation:rule=(self.tier != "REGIONAL" || size(self.fileShareName) <= 32), message="REGIONAL tier fileShareName length must be 32 or less characters"
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
