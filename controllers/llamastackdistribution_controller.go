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

package controllers

import (
	"context"
	"fmt"

	llamav1alpha1 "github.com/meta-llama/llama-stack-k8s-operator/api/v1alpha1"
	"github.com/meta-llama/llama-stack-k8s-operator/pkg/deploy"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defaultContainerName   = "llama-stack-server"
	defaultPort            = 8321 // Matches the QuickStart guide
	defaultServicePortName = "http"
	defaultLabelKey        = "app"
	defaultLabelValue      = "llama-stack"
)

// LlamaStackDistributionReconciler reconciles a LlamaStack object
type LlamaStackDistributionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Images map[string]string // Pre-defined image mappings for distributions
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LlamaStack object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *LlamaStackDistributionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("llamastack", req.NamespacedName)

	// Fetch the LlamaStack instance
	instance := &llamav1alpha1.LlamaStackDistribution{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("LlamaStack resource not found")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to fetch LlamaStack")
		return ctrl.Result{}, fmt.Errorf("fetching LlamaStack: %w", err)
	}

	// Reconcile the Deployment
	if err := r.reconcileDeployment(ctx, instance); err != nil {
		logger.Error(err, "Failed to reconcile Deployment")
		return ctrl.Result{}, fmt.Errorf("reconciling Deployment: %w", err)
	}

	// Reconcile the Service if ports are defined, else use default port
	if deploy.HasPorts(instance) {
		if err := r.reconcileService(ctx, instance); err != nil {
			logger.Error(err, "Failed to reconcile Service")
			return ctrl.Result{}, fmt.Errorf("reconciling Service: %w", err)
		}
	}

	// Update status
	if err := r.updateStatus(ctx, instance); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, fmt.Errorf("updating status: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LlamaStackDistributionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&llamav1alpha1.LlamaStackDistribution{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

// reconcileDeployment manages the Deployment for the LlamaStack server
func (r *LlamaStackDistributionReconciler) reconcileDeployment(ctx context.Context, instance *llamav1alpha1.LlamaStackDistribution) error {
	logger := log.FromContext(ctx)
	llamaServerImage := instance.Spec.Image

	// Prepare the pod spec with a default container if none provided
	podSpec := instance.Spec.Template
	if len(podSpec.Containers) == 0 {
		podSpec.Containers = []corev1.Container{{
			Name:  defaultContainerName,
			Image: llamaServerImage,
			Ports: []corev1.ContainerPort{{ContainerPort: defaultPort}},
		}}
	} else {
		for i, container := range podSpec.Containers {
			if container.Name == defaultContainerName {
				podSpec.Containers[i].Image = llamaServerImage
			}
		}
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &instance.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{defaultLabelKey: defaultLabelValue},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{defaultLabelKey: defaultLabelValue},
				},
				Spec: podSpec,
			},
		},
	}

	return deploy.ApplyDeployment(ctx, r.Client, r.Scheme, instance, deployment, logger)
}

// reconcileService manages the Service if ports are defined
func (r *LlamaStackDistributionReconciler) reconcileService(ctx context.Context, instance *llamav1alpha1.LlamaStackDistribution) error {
	logger := log.FromContext(ctx)

	// Use the first container's first port (consistent with guide)
	container := instance.Spec.Template.Containers[0]
	if len(container.Ports) == 0 {
		return nil // No ports to expose
	}
	port := container.Ports[0].ContainerPort

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-service",
			Namespace: instance.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{defaultLabelKey: defaultLabelValue},
			Ports: []corev1.ServicePort{{
				Name: defaultServicePortName,
				Port: port,
				TargetPort: intstr.IntOrString{
					IntVal: port,
				},
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	return deploy.ApplyService(ctx, r.Client, r.Scheme, instance, service, logger)
}

// updateStatus refreshes the LlamaStack status
func (r *LlamaStackDistributionReconciler) updateStatus(ctx context.Context, instance *llamav1alpha1.LlamaStackDistribution) error {
	logger := log.FromContext(ctx)

	instance.Status.Image = instance.Spec.Image

	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("fetching Deployment for status: %w", err)
	}
	expectedReplicas := instance.Spec.Replicas
	instance.Status.Ready = err == nil && deployment.Status.ReadyReplicas == expectedReplicas

	if err := r.Status().Update(ctx, instance); err != nil {
		logger.Error(err, "Failed to update LlamaStack status")
		return fmt.Errorf("updating status: %w", err)
	}
	return nil
}

// NewLlamaStackDistributionReconciler creates a new reconciler with default image mappings
func NewLlamaStackDistributionReconciler(client client.Client, scheme *runtime.Scheme) *LlamaStackDistributionReconciler {
	return &LlamaStackDistributionReconciler{
		Client: client,
		Scheme: scheme,
	}
}
