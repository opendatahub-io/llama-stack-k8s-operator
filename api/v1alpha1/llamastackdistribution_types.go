/*
Copyright 2025.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LlamaStackDistributionSpec defines the desired state of LlamaStackDistribution
type LlamaStackDistributionSpec struct {
	Replicas *int32         `json:"replicas,omitempty"`
	Image    string         `json:"image"`
	Template corev1.PodSpec `json:"template,omitempty"`
}

// LlamaStackDistributionStatus defines the observed state of LlamaStackDistribution
type LlamaStackDistributionStatus struct {
	Image string `json:"image,omitempty"`
	Ready bool   `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Distribution",type="string",JSONPath=".status.distribution"
//+kubebuilder:printcolumn:name="Image",type="string",JSONPath=".status.image"
//+kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// LlamaStackDistribution is the Schema for the llamastackdistributions API

type LlamaStackDistribution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LlamaStackDistributionSpec   `json:"spec"`
	Status LlamaStackDistributionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LlamaStackDistributionList contains a list of LlamaStackDistribution
type LlamaStackDistributionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LlamaStackDistribution `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LlamaStackDistribution{}, &LlamaStackDistributionList{})
}
