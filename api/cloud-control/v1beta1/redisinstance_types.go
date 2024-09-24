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

type TransitEncryptionGcp struct {
	// Client to Server traffic encryption enabled with server authentication.
	// +optional
	// +kubebuilder:default=false
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="ServerAuthentication is immutable."
	ServerAuthentication bool `json:"serverAuthentication,omitempty"`
}

type TimeOfDayGcp struct {
	// Hours of day in 24 hour format. Should be from 0 to 23.
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

type DayOfWeekPolicyGcp struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=MONDAY;TUESDAY;WEDNESDAY;THURSDAY;FRIDAY;SATURDAY;SUNDAY;
	Day string `json:"day"`

	// +kubebuilder:validation:Required
	StartTime TimeOfDayGcp `json:"startTime"`
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type MaintenancePolicyGcp struct {
	DayOfWeek *DayOfWeekPolicyGcp `json:"dayOfWeek,omitempty"`
}

type RedisInstanceGcp struct {
	// The service tier of the instance.
	// +kubebuilder:default=BASIC
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Tier is immutable."
	// +kubebuilder:validation:Enum=BASIC;STANDARD_HA
	Tier string `json:"tier"`

	// Redis memory size in GiB.
	// +kubebuilder:validation:Required
	MemorySizeGb int32 `json:"memorySizeGb"`

	// The version of Redis software.
	// +optional
	// +kubebuilder:default=REDIS_7_0
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RedisVersion is immutable."
	// +kubebuilder:validation:Enum=REDIS_7_2;REDIS_7_0;REDIS_6_X;REDIS_5_0;REDIS_4_0;REDIS_3_2
	RedisVersion string `json:"redisVersion"`

	// Indicates whether OSS Redis AUTH is enabled for the instance.
	// +optional
	// +kubebuilder:default=false
	AuthEnabled bool `json:"authEnabled"`

	// The TLS mode of the Redis instance.
	// If not provided, TLS is disabled for the instance.
	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="TransitEncryption is immutable."
	TransitEncryption *TransitEncryptionGcp `json:"transitEncryption,omitempty"`

	// Redis configuration parameters, according to http://redis.io/topics/config.
	// See docs for the list of the supported parameters
	// +optional
	RedisConfigs map[string]string `json:"redisConfigs"`

	// The maintenance policy for the instance.
	// If not provided, maintenance events can be performed at any time.
	// +optional
	MaintenancePolicy *MaintenancePolicyGcp `json:"maintenancePolicy,omitempty"`
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

// +kubebuilder:validation:XValidation:rule=(self.authEnabled == false || self.transitEncryptionEnabled == true), message="authEnabled can only be true if TransitEncryptionEnabled is also true"
type RedisInstanceAws struct {
	// +kubebuilder:validation:Required
	CacheNodeType string `json:"cacheNodeType"`

	// +optional
	// +kubebuilder:default="7.0"
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="EngineVersion is immutable."
	EngineVersion string `json:"engineVersion"`

	// +optional
	// +kubebuilder:default=false
	AutoMinorVersionUpgrade bool `json:"autoMinorVersionUpgrade"`

	// +optional
	// +kubebuilder:default=false
	TransitEncryptionEnabled bool `json:"transitEncryptionEnabled"`

	// +optional
	// +kubebuilder:default=false
	AuthEnabled bool `json:"authEnabled"`

	// Specifies the weekly time range during which maintenance on the cluster is
	// performed. It is specified as a range in the format ddd:hh24:mi-ddd:hh24:mi (24H
	// Clock UTC). The minimum maintenance window is a 60 minute period.
	//
	// Valid values for ddd are: sun mon tue wed thu fri sat
	//
	// Example: sun:23:00-mon:01:30
	// +optional
	PreferredMaintenanceWindow *string `json:"preferredMaintenanceWindow,omitempty"`

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
