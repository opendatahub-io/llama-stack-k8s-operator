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

// DistributionType defines the distribution configuration for llama-stack.
type DistributionType struct {
	// Name is the distribution name that maps to a predefined image in imageMap
	// +optional
	Name string `json:"name,omitempty"`
	// Image is the direct container image reference to use
	// +optional
	Image string `json:"image,omitempty"`
}

// LlamaStackDistributionSpec defines the desired state of LlamaStackDistribution.
type LlamaStackDistributionSpec struct {
	// +kubebuilder:default:=1
	Replicas int32      `json:"replicas,omitempty"`
	Server   ServerSpec `json:"server"`
}

// ServerSpec defines the desired state of llama server.
type ServerSpec struct {
	// +kubebuilder:default:={"name": "ollama"}
	Distribution  DistributionType `json:"distribution"`
	ContainerSpec ContainerSpec    `json:"containerSpec"`
	PodOverrides  *PodOverrides    `json:"podOverrides,omitempty"` // Optional pod-level overrides
}

// ContainerSpec defines the llama-stack server container configuration.
type ContainerSpec struct {
	// +kubebuilder:default:="llama-stack"
	Name      string                      `json:"name,omitempty"` // Optional, defaults to "llama-stack"
	Port      int32                       `json:"port,omitempty"` // Defaults to 8321 if unset
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	Env       []corev1.EnvVar             `json:"env,omitempty"` // Runtime env vars (e.g., INFERENCE_MODEL)
}

// PodOverrides allows advanced pod-level customization.
type PodOverrides struct {
	Volumes      []corev1.Volume      `json:"volumes,omitempty"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// ProviderInfo represents a single provider from the providers endpoint.
type ProviderInfo struct {
	API          string `json:"api"`
	ProviderID   string `json:"provider_id"`
	ProviderType string `json:"provider_type"`
}

// DistributionConfig represents the configuration information from the providers endpoint.
type DistributionConfig struct {
	Providers []ProviderInfo `json:"providers,omitempty"`
}

// LlamaStackDistributionStatus defines the observed state of LlamaStackDistribution.
type LlamaStackDistributionStatus struct {
	Version            string             `json:"version,omitempty"`
	DistributionConfig DistributionConfig `json:"distributionConfig,omitempty"`
	Ready              bool               `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"
//+kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// LlamaStackDistribution is the Schema for the llamastackdistributions API

type LlamaStackDistribution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LlamaStackDistributionSpec   `json:"spec"`
	Status LlamaStackDistributionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LlamaStackDistributionList contains a list of LlamaStackDistribution.
type LlamaStackDistributionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LlamaStackDistribution `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&LlamaStackDistribution{}, &LlamaStackDistributionList{})
}

// HasPorts checks if the container spec defines a port.
func (r *LlamaStackDistribution) HasPorts() bool {
	return r.Spec.Server.ContainerSpec.Port != 0 || len(r.Spec.Server.ContainerSpec.Env) > 0 // Port or env implies service need
}
