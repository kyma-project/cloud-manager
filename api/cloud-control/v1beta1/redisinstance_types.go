/*
Copyright 2023.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	armRedis "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	ReasonCanNotLoadResourceGroup   = "ResourceGroupCanNotLoad"
	ReasonCanNotDeleteResourceGroup = "ResourceGroupCanNotDelete"

	ReasonCanNotCreateResourceGroup = "ResourceGroupCanNotCreate"
)

// RedisInstanceSpec defines the desired state of RedisInstance
type RedisInstanceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteRef is immutable."
	RemoteRef RemoteRef `json:"remoteRef"`

	// +kubebuilder:validation:Required
	IpRange IpRangeRef `json:"ipRange"`

	// +kubebuilder:validation:Required
	Scope ScopeRef `json:"scope"`

	// +kubebuilder:validation:Required
	Instance RedisInstanceInfo `json:"instance"`
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type RedisInstanceInfo struct {
	// +optional
	Gcp *RedisInstanceGcp `json:"gcp,omitempty"`

	// +optional
	Azure *RedisInstanceAzure `json:"azure,omitempty"`

	// +optional
	Aws *RedisInstanceAws `json:"aws,omitempty"`
}

type RedisInstanceAzureConfigs struct {
	// +optional
	MaxClients string `json:"maxclients,omitempty"`
	// +optional
	MaxFragmentationMemoryReserved string `json:"maxfragmentationmemory-reserved,omitempty"`
	// +optional
	MaxMemoryDelta string `json:"maxmemory-delta,omitempty"`
	// +optional
	MaxMemoryPolicy string `json:"maxmemory-policy,omitempty"`
	// +optional
	MaxMemoryReserved string `json:"maxmemory-reserved,omitempty"`
	// +optional
	NotifyKeyspaceEvents string `json:"notify-keyspace-events,omitempty"`
	// +optional
	ZonalConfiguration string `json:"zonal-configuration,omitempty"`
}

func (redisConfigs *RedisInstanceAzureConfigs) GetRedisConfig() *armRedis.CommonPropertiesRedisConfiguration {
	redisConfiguration := armRedis.CommonPropertiesRedisConfiguration{}

	additionalProperties := map[string]interface{}{}

	if redisConfigs.MaxFragmentationMemoryReserved != "" {
		redisConfiguration.MaxfragmentationmemoryReserved = &redisConfigs.MaxFragmentationMemoryReserved
	}
	if redisConfigs.MaxMemoryDelta != "" {
		redisConfiguration.MaxmemoryDelta = &redisConfigs.MaxMemoryDelta
	}
	if redisConfigs.MaxMemoryPolicy != "" {
		redisConfiguration.MaxmemoryPolicy = &redisConfigs.MaxMemoryPolicy
	}
	if redisConfigs.MaxMemoryReserved != "" {
		redisConfiguration.MaxmemoryReserved = &redisConfigs.MaxMemoryReserved
	}
	if redisConfigs.NotifyKeyspaceEvents != "" {
		additionalProperties["notify-keyspace-events"] = &redisConfigs.NotifyKeyspaceEvents
	}
	if redisConfigs.MaxClients != "" {
		redisConfiguration.Maxclients = &redisConfigs.MaxClients
	}
	if redisConfigs.ZonalConfiguration != "" {
		redisConfiguration.ZonalConfiguration = &redisConfigs.ZonalConfiguration
	}

	if len(additionalProperties) > 0 {
		redisConfiguration.AdditionalProperties = additionalProperties
	}

	return &redisConfiguration
}

type AzureRedisSKU struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=1;2;3;4
	Capacity int `json:"capacity"`
}

type TimeOfDayGcp struct {
	// Hours of day in 24 hour format. Should be from 0 to 23. An API may choose
	// to allow the value "24:00:00" for scenarios like business closing time.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=23
	Hours int32 `json:"hours"`

	// Minutes of hour of day. Must be from 0 to 59.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=59
	Minutes int32 `json:"minutes"`
}

type WeeklyMaintenanceWindowGcp struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=DAY_OF_WEEK_UNSPECIFIED;MONDAY;TUESDAY;WEDNESDAY;THURSDAY;FRIDAY;SATURDAY;SUNDAY;
	Day string `json:"day"`

	// Start time of the window in UTC time.
	// +kubebuilder:validation:Required
	StartTime TimeOfDayGcp `json:"startTime"`
}

type RedisInstanceGcp struct {
	// +kubebuilder:default=BASIC
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Tier is immutable."
	// +kubebuilder:validation:Enum=BASIC;STANDARD_HA
	Tier string `json:"tier"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="MemorySizeGb is immutable."
	MemorySizeGb int32 `json:"memorySizeGb"`

	// +optional
	// +kubebuilder:default=REDIS_7_0
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RedisVersion is immutable."
	// +kubebuilder:validation:Enum=REDIS_7_0;REDIS_6_X;REDIS_5_0;REDIS_4_0;REDIS_3_2
	RedisVersion string `json:"redisVersion"`

	// +optional
	// +kubebuilder:default=true
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="AuthEnabled is immutable."
	AuthEnabled bool `json:"authEnabled,omitempty"`

	// +optional
	// +kubebuilder:default=TRANSIT_ENCRYPTION_MODE_UNSPECIFIED
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="TransitEncryptionMode is immutable."
	// +kubebuilder:validation:Enum=TRANSIT_ENCRYPTION_MODE_UNSPECIFIED;SERVER_AUTHENTICATION;DISABLED
	TransitEncryptionMode string `json:"transitEncryptionMode,omitempty"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RedisConfigs is immutable."
	RedisConfigs map[string]string `json:"redisConfigs,omitempty"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="MaintenancePolicy is immutable."
	MaintenancePolicy *WeeklyMaintenanceWindowGcp `json:"maintenancePolicy,omitempty"`
}

type RedisInstanceAzure struct {
	// +kubebuilder:validation:Required
	SKU AzureRedisSKU `json:"sku"`

	// +optional
	EnableNonSslPort bool `json:"enableNonSslPort,omitempty"`

	// +optional
	RedisConfiguration RedisInstanceAzureConfigs `json:"redisConfiguration"`

	// +optional
	RedisVersion string `json:"redisVersion,omitempty"`

	// +optional
	ShardCount int `json:"shardCount,omitempty"`

	// +optional
	ReplicasPerPrimary int `json:"replicasPerPrimary,omitempty"`
}

type RedisInstanceAws struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="CacheNodeType is immutable."
	CacheNodeType string `json:"cacheNodeType"`

	// +optional
	// +kubebuilder:default="7.0"
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="EngineVersion is immutable."
	EngineVersion string `json:"engineVersion"`

	// +optional
	// +kubebuilder:default=false
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="AutoMinorVersionUpgrade is immutable."
	AutoMinorVersionUpgrade bool `json:"autoMinorVersionUpgrade"`

	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`
}

// RedisInstanceStatus defines the observed state of RedisInstance
type RedisInstanceStatus struct {
	State StatusState `json:"state,omitempty"`

	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	PrimaryEndpoint string `json:"primaryEndpoint,omitempty"`

	// +optional
	ReadEndpoint string `json:"readEndpoint,omitempty"`

	// +optional
	AuthString string `json:"authString,omitempty"`

	// +optional
	CaCert string `json:"caCert,omitempty"`

	// List of status conditions to indicate the status of a RedisInstance.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RedisInstance is the Schema for the redisinstances API
type RedisInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisInstanceSpec   `json:"spec,omitempty"`
	Status RedisInstanceStatus `json:"status,omitempty"`
}

func (in *RedisInstance) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *RedisInstance) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *RedisInstance) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *RedisInstance) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *RedisInstance) SetStatusStateToReady() {
	in.Status.State = ReadyState
}

func (in *RedisInstance) SetStatusStateToError() {
	in.Status.State = ErrorState
}

//+kubebuilder:object:root=true

// RedisInstanceList contains a list of RedisInstance
type RedisInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisInstance{}, &RedisInstanceList{})
}
