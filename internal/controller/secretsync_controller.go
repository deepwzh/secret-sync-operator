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

package controller

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	syncv1 "github.com/deepwzh/secret-sync-operator/api/v1"
)

// SecretSyncReconciler reconciles a SecretSync object
type SecretSyncReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Clientset *kubernetes.Clientset
}

// +kubebuilder:rbac:groups=sync.92ac.cn,resources=secretsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sync.92ac.cn,resources=secretsyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=sync.92ac.cn,resources=secretsyncs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretSync object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *SecretSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the SecretSync instance
	secretSync := &syncv1.SecretSync{}
	err := r.Get(ctx, req.NamespacedName, secretSync)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get the source secret
	secret := &corev1.Secret{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: secretSync.Spec.SecretName, Namespace: req.Namespace}, secret)
	if err != nil {
		logger.Error(err, "Failed to get source secret")
		return ctrl.Result{}, err
	}

	// Determine target namespaces
	var targetNamespaces []string
	if secretSync.Spec.Namespaces == "*" {
		nsList, err := r.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Error(err, "Failed to list namespaces")
			return ctrl.Result{}, err
		}
		for _, ns := range nsList.Items {
			targetNamespaces = append(targetNamespaces, ns.Name)
		}
	} else {
		targetNamespaces = strings.Split(secretSync.Spec.Namespaces, ",")
	}

	// Sync secret to target namespaces
	for _, ns := range targetNamespaces {
		targetSecret := &corev1.Secret{}
		err := r.Client.Get(ctx, types.NamespacedName{Name: secretSync.Spec.SecretName, Namespace: ns}, targetSecret)
		if err != nil && errors.IsNotFound(err) {
			// Secret not found, create it
			newSecret := secret.DeepCopy()
			newSecret.Namespace = ns
			newSecret.ResourceVersion = ""
			err = r.Client.Create(ctx, newSecret)
			if err != nil {
				logger.Error(err, "Failed to create secret in namespace", "namespace", ns)
				return ctrl.Result{}, err
			}
			logger.Info("Secret created in namespace", "secret", secretSync.Spec.SecretName, "source_namespace", req.Namespace, "target_namespace", ns)
		} else if err != nil {
			logger.Error(err, "Failed to get secret in namespace", "namespace", ns)
			return ctrl.Result{}, err
		}
	}

	// Update status
	secretSync.Status.SyncedNamespaces = targetNamespaces
	err = r.Status().Update(ctx, secretSync)
	if err != nil {
		logger.Error(err, "Failed to update SecretSync status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	config := ctrl.GetConfigOrDie()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	r.Clientset = clientset

	return ctrl.NewControllerManagedBy(mgr).
		For(&syncv1.SecretSync{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}
