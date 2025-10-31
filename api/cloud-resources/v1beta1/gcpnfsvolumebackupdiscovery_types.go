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
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GcpNfsVolumeBackupDiscoverySpec defines the desired state of GcpNfsVolumeBackupDiscovery
type GcpNfsVolumeBackupDiscoverySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// GcpNfsVolumeBackupDiscoveryStatus defines the observed state of GcpNfsVolumeBackupDiscovery
type GcpNfsVolumeBackupDiscoveryStatus struct {
	// +optional
	State string `json:"state,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	DiscoverySnapshotTime *metav1.Time `json:"discoverySnapshotTime,omitempty"`

	// +optional
	AvailableBackupsCount *int `json:"availableBackupsCount,omitempty"`

	// +optional
	AvailableBackupUris []string `json:"availableBackupUris,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="AvailableBackupsCount",type="integer",JSONPath=".status.availableBackupsCount"
// +kubebuilder:printcolumn:name="DiscoverySnapshotTime",type="date",JSONPath=".status.discoverySnapshotTime"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// GcpNfsVolumeBackupDiscovery is the Schema for the gcpnfsvolumebackupdiscoveries API
type GcpNfsVolumeBackupDiscovery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpNfsVolumeBackupDiscoverySpec   `json:"spec,omitempty"`
	Status GcpNfsVolumeBackupDiscoveryStatus `json:"status,omitempty"`
}

func (in *GcpNfsVolumeBackupDiscovery) State() string {
	return string(in.Status.State)
}

func (in *GcpNfsVolumeBackupDiscovery) SetState(v string) {
	in.Status.State = v
}

func (in *GcpNfsVolumeBackupDiscovery) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *GcpNfsVolumeBackupDiscovery) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *GcpNfsVolumeBackupDiscovery) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (in *GcpNfsVolumeBackupDiscovery) SpecificToProviders() []string {
	return []string{"gcp"}
}

//+kubebuilder:object:root=true

// GcpNfsVolumeBackupDiscoveryList contains a list of GcpNfsVolumeBackupDiscovery
type GcpNfsVolumeBackupDiscoveryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpNfsVolumeBackupDiscovery `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpNfsVolumeBackupDiscovery{}, &GcpNfsVolumeBackupDiscoveryList{})
}

func (in *GcpNfsVolumeBackupDiscovery) CloneForPatchStatus() client.Object {
	return &GcpNfsVolumeBackupDiscovery{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GcpNfsVolumeBackupDiscovery",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}
