/*
Copyright 2024 maniraja1122.

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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"slices"
	"strings"
)

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Additional Functions
func (r *SecretReconciler) deleteSecret(ctx context.Context, name, namespace string) error {
	// Define the Secret to delete
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	// Delete the Secret
	if err := r.Delete(ctx, secret); err != nil {
		return err // Handle the error appropriately
	}

	return nil
}
func (r *SecretReconciler) NamespaceExists(ctx context.Context, namespaceName string) (bool, error) {
	ns := &corev1.Namespace{}
	err := r.Get(ctx, types.NamespacedName{Name: namespaceName}, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil // Namespace does not exist
		}
		return false, err // An error occurred while trying to get the namespace
	}
	return true, nil // Namespace exists
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Secret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	secret := &corev1.Secret{}
	err := r.Get(ctx, req.NamespacedName, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, err.Error())
		return ctrl.Result{}, err
	}
	// Checking Potential Deletion
	if secret.DeletionTimestamp != nil {
		duplicated_ns_list, list_exist := secret.Annotations["datareplicator/replicate-to"]
		if list_exist {
			ns_list := strings.Split(duplicated_ns_list, ",")
			for _, n := range ns_list {
				err := r.deleteSecret(ctx, secret.Name, n)
				if err != nil {
					return ctrl.Result{}, err
				}
			}
		}
		secret.Finalizers = slices.DeleteFunc(secret.Finalizers, func(v string) bool {
			if v == finalizerMessage {
				return true
			} else {
				return false
			}
		})
		err := r.Update(ctx, secret)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	// Duplicating....
	labels := secret.Annotations
	currentnamespace := secret.Namespace
	val, exist := labels["datareplicator/replicate-to"]
	if exist {
		// Make List of Namespaces, Remove Duplicates
		namespaces := strings.Split(val, ",")
		slices.Sort(namespaces)
		namespaces = slices.Compact(namespaces)
		for _, n := range namespaces {
			if n != currentnamespace {
				namespaceExist, err := r.NamespaceExists(ctx, n)
				if err != nil {
					logger.Error(err, err.Error())
					return ctrl.Result{}, err
				}
				if !namespaceExist {
					namespaceRequested, requestFound := labels["datareplicator/createnamespace"]
					if requestFound && namespaceRequested == "true" {
						newNamespace := &corev1.Namespace{
							ObjectMeta: metav1.ObjectMeta{
								Name: n,
							},
						}
						err := r.Create(ctx, newNamespace)
						if err != nil {
							logger.Error(err, err.Error())
							return ctrl.Result{}, err
						}
					} else {
						continue
					}
				}
				// Proceed with Creation of Duplicate
				copy_secret := secret.DeepCopy()
				// Remove Identity Field
				copy_secret.UID = ""
				copy_secret.ResourceVersion = ""
				// Change Namespace
				copy_secret.Namespace = n
				// Modifying labels for duplicate
				copy_secret.Annotations["datareplicator/sourcenamespace"] = currentnamespace
				delete(copy_secret.Annotations, "datareplicator/replicate-to")
				delete(copy_secret.Annotations, "datareplicator/createnamespace")
				// Update Current or Add New
				find_secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secret.Name,
						Namespace: n,
					},
				}
				err = r.Get(ctx, types.NamespacedName{
					Namespace: n,
					Name:      secret.Name,
				}, find_secret)
				if err != nil {
					if errors.IsNotFound(err) {
						logger.Info("Creating new Secret")
						// Modifying labels for Original
						secret.Annotations["datareplicator/replicated"] = "true"
						// Launch New and Update [Add Finalizer]
						secret.Finalizers = append(secret.Finalizers, finalizerMessage)
						err = r.Update(ctx, secret)
						if err != nil {
							logger.Error(err, err.Error())
							return ctrl.Result{}, err
						}
						err = r.Create(ctx, copy_secret)
						if err != nil {
							logger.Error(err, err.Error())
							return ctrl.Result{}, err
						}
					} else {
						logger.Error(err, err.Error())
						return ctrl.Result{}, err
					}
				} else {
					logger.Info("Modifying Old Secret")
					find_secret.Data = secret.Data
					err = r.Update(ctx, find_secret)
					if err != nil {
						logger.Error(err, err.Error())
						return ctrl.Result{}, err
					}
				}
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}
