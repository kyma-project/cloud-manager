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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// CertificateSpec defines the desired state of Certificate
type CertificateSpec struct {
	// SecretRef references a Secret containing certificate data
	// Required keys: "tls.crt", "tls.key"
	// Optional keys: "ca.crt" (certificate chain)
	// +kubebuilder:validation:Required
	SecretRef klog.ObjectRef `json:"secretRef"`
}

// CertificateStatus defines the observed state of Certificate
type CertificateStatus struct {
	// Id is the Cloud Manager generated unique identifier for the cloud resource,
	// used as the cloud resource name to avoid collision across namespaces and SKRs.
	// +optional
	Id string `json:"id,omitempty"`

	// ProviderId is the provider-specific resource identifier (e.g., AWS ARN, Azure Resource ID)
	// +optional
	ProviderId string `json:"providerId,omitempty"`

	// List of status conditions to indicate the status of a Certificate
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="ProviderID",type="string",JSONPath=".status.providerId"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Certificate is the Schema for the certificates API
type Certificate struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of Certificate
	// +required
	Spec CertificateSpec `json:"spec"`

	// status defines the observed state of Certificate
	// +optional
	Status CertificateStatus `json:"status,omitempty"`
}

func (in *Certificate) Id() string     { return in.Status.Id }
func (in *Certificate) SetId(v string) { in.Status.Id = v }

func (in *Certificate) ObservedGeneration() int64 {
	readyCondition := meta.FindStatusCondition(in.Status.Conditions, ConditionTypeReady)
	if readyCondition == nil {
		return 0
	}
	return readyCondition.ObservedGeneration
}

func (in *Certificate) SetObservedGeneration(i int64) {
	// ObservedGeneration is managed through the Ready condition
}

func (in *Certificate) GetStatus() any {
	return &in.Status
}

func (in *Certificate) SetStatusProviderError(msg string) {
	in.Status.State = ReasonProviderError
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonProviderError,
		Message:            msg,
	})
}

func (in *Certificate) SetStatusReady() {
	in.Status.State = StateReady
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: in.Generation,
		Reason:             ReasonAvailable,
		Message:            ReasonAvailable,
	})
}

func (in *Certificate) SetStatusProcessing() {
	in.Status.State = StateProcessing
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionUnknown,
		ObservedGeneration: in.Generation,
		Reason:             ReasonProcessing,
		Message:            ReasonProcessing,
	})
}

func (in *Certificate) SetStatusDeleting() {
	in.Status.State = StateDeleting
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionUnknown,
		ObservedGeneration: in.Generation,
		Reason:             ReasonDeleting,
		Message:            ReasonDeleting,
	})
}

func (in *Certificate) SetStatusDeletionBlockedByDependents(msg string) {
	in.Status.State = ReasonDeletionBlockedByDependents
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonDeletionBlockedByDependents,
		Message:            msg,
	})
}

func (in *Certificate) RemoveStatusDeletionBlockedByDependents() {
	in.SetStatusDeleting()
}

func (in *Certificate) Conditions() *[]metav1.Condition { return &in.Status.Conditions }

func (in *Certificate) GetObjectMeta() *metav1.ObjectMeta { return &in.ObjectMeta }

func (in *Certificate) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureCertificate
}

func (in *Certificate) SpecificToProviders() []string { return []string{"aws"} }

func (in *Certificate) State() string {
	return in.Status.State
}

func (in *Certificate) SetState(v string) {
	in.Status.State = v
}

// +kubebuilder:object:root=true

// CertificateList contains a list of Certificate
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Certificate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Certificate{}, &CertificateList{})
}
