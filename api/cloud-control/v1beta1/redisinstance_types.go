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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

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

type RedisInstanceGcpConfigs struct {
	// +optional
	MaxmemoryPolicy string `json:"maxmemory-policy,omitempty"`
	// +optional
	NotifyKeyspaceEvents string `json:"notify-keyspace-events,omitempty"`

	// +optional
	Activedefrag string `json:"activedefrag,omitempty"`
	// +optional
	LfuDecayTime string `json:"lfu-decay-time,omitempty"`
	// +optional
	LfuLogFactor string `json:"lfu-log-factor,omitempty"`
	// +optional
	MaxmemoryGb string `json:"maxmemory-gb,omitempty"`

	// +optional
	StreamNodeMaxBytes string `json:"stream-node-max-bytes,omitempty"`
	// +optional
	StreamNodeMaxEntries string `json:"stream-node-max-entries,omitempty"`
}

type RedisInstanceAzureConfigs struct {
	// +optional
	AadEnabled string `json:"aad-enabled,omitempty"`
	// +optional
	AofBackupEnabled string `json:"aof-backup-enabled,omitempty"`
	// +optional
	AofStorageConnectionString0 string `json:"aof-storage-connection-string-0,omitempty"`
	// +optional
	AofStorageConnectionString1 string `json:"aof-storage-connection-string-1,omitempty"`
	// +optional
	AuthNotRequired string `json:"authnotrequired,omitempty"`
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
	PreferredDataArchiveAuthMethod string `json:"preferred-data-archive-auth-method,omitempty"`
	// +optional
	PreferredDataPersistenceAuthMethod string `json:"preferred-data-persistence-auth-method,omitempty"`
	// +optional
	RdbBackupEnabled string `json:"rdb-backup-enabled,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum=15;30;60;360;720;1440
	RdbBackupFrequency string `json:"rdb-backup-frequency,omitempty"`
	// +optional
	RdbBackupMaxSnapshotCount string `json:"rdb-backup-max-snapshot-count,omitempty"`
	// +optional
	RdbStorageConnectionString string `json:"rdb-storage-connection-string,omitempty"`
	// +optional
	StorageSubscriptionId string `json:"storage-subscription-id,omitempty"`
	// +optional
	ZonalConfiguration string `json:"zonal-configuration,omitempty"`
}

type AzureRedisSKU struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Basic;Standard;Premium
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=C;P
	Family string `json:"family"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=0;1;2;3;4;5;6
	Capacity int `json:"capacity"`
}

type RedisInstanceAzureProperties struct {
	// +kubebuilder:validation:Required
	SKU AzureRedisSKU `json:"sku,omitempty"`

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

func (redisConfigs *RedisInstanceGcpConfigs) ToMap() map[string]string {
	result := map[string]string{}

	if redisConfigs.MaxmemoryPolicy != "" {
		result["maxmemory-policy"] = redisConfigs.MaxmemoryPolicy
	}
	if redisConfigs.NotifyKeyspaceEvents != "" {
		result["notify-keyspace-events"] = redisConfigs.NotifyKeyspaceEvents
	}

	if redisConfigs.Activedefrag != "" {
		result["activedefrag"] = redisConfigs.Activedefrag
	}
	if redisConfigs.LfuDecayTime != "" {
		result["lfu-decay-time"] = redisConfigs.LfuDecayTime
	}
	if redisConfigs.LfuLogFactor != "" {
		result["lfu-log-factor"] = redisConfigs.LfuLogFactor
	}
	if redisConfigs.MaxmemoryGb != "" {
		result["maxmemory-gb"] = redisConfigs.MaxmemoryGb
	}

	if redisConfigs.StreamNodeMaxBytes != "" {
		result["stream-node-max-bytes"] = redisConfigs.StreamNodeMaxBytes
	}
	if redisConfigs.StreamNodeMaxEntries != "" {
		result["stream-node-max-entries"] = redisConfigs.StreamNodeMaxEntries
	}

	return result
}

type RedisInstanceGcp struct {
	// +kubebuilder:default=BASIC
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Tier is immutable."
	Tier string `json:"tier"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="MemorySizeGb is immutable."
	MemorySizeGb int32 `json:"memorySizeGb"`

	// +kubebuilder:default=REDIS_7_0
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RedisVersion is immutable."
	// +kubebuilder:validation:Enum=REDIS_7_0;REDIS_6_X;REDIS_5_0;REDIS_4_0;REDIS_3_2
	RedisVersion string `json:"redisVersion"`

	// +optional
	// +kubebuilder:default=true
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="AuthEnabled is immutable."
	AuthEnabled bool `json:"authEnabled"`

	// +optional
	// +kubebuilder:default=TRANSIT_ENCRYPTION_MODE_UNSPECIFIED
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="TransitEncryptionMode is immutable."
	// +kubebuilder:validation:Enum=TRANSIT_ENCRYPTION_MODE_UNSPECIFIED;SERVER_AUTHENTICATION;DISABLED
	TransitEncryptionMode string `json:"transitEncryptionMode"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RedisConfigs is immutable."
	RedisConfigs RedisInstanceGcpConfigs `json:"redisConfigs"`
}

type RedisInstanceAzure struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Name is immutable."
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	SubscriptionId string `json:"subscriptionId"`

	// +kubebuilder:validation:Required
	ResourceGroupName string `json:"resourceGroupName"`

	// +kubebuilder:validation:Required
	ApiVersion string `json:"apiVersion"`

	// +kubebuilder:validation:Required
	Properties RedisInstanceAzureProperties `json:"properties"`
}

type RedisInstanceAws struct {
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
