package deploy

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	llamav1alpha1 "github.com/meta-llama/llama-stack-k8s-operator/api/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplyNetworkPolicy creates or updates a NetworkPolicy.
func ApplyNetworkPolicy(ctx context.Context, c client.Client, scheme *runtime.Scheme,
	instance *llamav1alpha1.LlamaStackDistribution, networkPolicy *networkingv1.NetworkPolicy, log logr.Logger) error {
	// Set the controller reference
	if err := ctrl.SetControllerReference(instance, networkPolicy, scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	// Check if the NetworkPolicy already exists
	existing := &networkingv1.NetworkPolicy{}
	err := c.Get(ctx, client.ObjectKeyFromObject(networkPolicy), existing)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create the NetworkPolicy if it doesn't exist
			if err := c.Create(ctx, networkPolicy); err != nil {
				return fmt.Errorf("failed to create NetworkPolicy: %w", err)
			}
			log.Info("Created NetworkPolicy", "name", networkPolicy.Name)
			return nil
		}
		return fmt.Errorf("failed to get NetworkPolicy: %w", err)
	}

	// Update the NetworkPolicy if it exists
	networkPolicy.ResourceVersion = existing.ResourceVersion
	if err := c.Update(ctx, networkPolicy); err != nil {
		return fmt.Errorf("failed to update NetworkPolicy: %w", err)
	}
	log.Info("Updated NetworkPolicy", "name", networkPolicy.Name)
	return nil
}

func GetOperatorNamespace() (string, error) {
	operatorNS, exist := os.LookupEnv("OPERATOR_NAMESPACE")
	if exist && operatorNS != "" {
		return operatorNS, nil
	}
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	return string(data), err
}
