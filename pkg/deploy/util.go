package deploy

import llamav1alpha1 "github.com/meta-llama/llama-stack-k8s-operator/api/v1alpha1"

// HasPorts checks if the template defines any container ports
func HasPorts(instance *llamav1alpha1.LlamaStackDistribution) bool {
	for _, container := range instance.Spec.Template.Containers {
		if len(container.Ports) > 0 {
			return true
		}
	}
	return false
}
