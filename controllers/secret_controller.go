/*


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
	"crypto/rand"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get;update;patch

// Reconcile is a reconciler for the core/v1 type Secret
func (r *SecretReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("secret", req.NamespacedName)

	// your logic here
	var secret corev1.Secret
	if err := r.Get(ctx, req.NamespacedName, &secret); err != nil {
		log.Error(err, "unable to get secrets", "secret", secret)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if _, err := r.patchSecret(secret); err != nil {
		log.Error(err, "unable to patch secret with new value", "secret", secret)
	}

	return ctrl.Result{}, nil

}

// compareTime is a function that evaluates whether a secret is more than 7 days old
// in such cases the secret that is more than 7 days old is returned to be used with secretGenerator
func (r *SecretReconciler) compareTime(secrets corev1.Secret) (secret corev1.Secret, notValid bool, err error) {
	secretTime := secrets.CreationTimestamp.Time
	targetTime := time.Now().AddDate(0, 0, -7)

	isNotValid := secretTime.Before(targetTime)
	//if isNotValid {
	//	return corev1.Secret{}, fmt.Errorf("Secret has expired: %v. Triggering new secret generation", corev1.Secret{})
	//}
	return secret, isNotValid, err
}

func (r *SecretReconciler) filterSecret(secret corev1.Secret) corev1.Secret {
	secret, notValid, err := r.compareTime(secret)
	if err != nil {
		fmt.Printf("error comparing the creation timestamp with target time for secret: %v", secret)
	}

	if notValid {
		return corev1.Secret{}
	}
	fmt.Println("expired secret identified")
	return secret
}

// secretGenerator is a function that will eventually take an int (which is the length of the secret to be generated)
// This will then be used within patchSecret
func (r *SecretReconciler) secretGenerator(s corev1.Secret) (corev1.Secret, map[string][]byte, error) {
	filtered := r.filterSecret(s)

	//logic for generating a random string which is used as the secret value
	value, err := r.generateRandomBytes(20)
	if err != nil {
		log.Errorf("Error creating new secret value for secret: %v", filtered)
	}

	log.Info("new secret value has been generated for expired secret")
	return filtered, value, err
}

func (r *SecretReconciler) patchSecret(s corev1.Secret) (corev1.Secret, error) {
	ctx := context.TODO()
	secretToPatch, newSecretVal, err := r.secretGenerator(s)

	//patch logic to be generated for patching the object
	patch := client.MergeFrom(s.DeepCopy())
	s.Data = newSecretVal
	r.Patch(ctx, &secretToPatch, patch)

	return secretToPatch, err
}

func (r *SecretReconciler) generateRandomBytes(n int) (map[string][]byte, error) {

	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	//logic to map []byte to map[string][]byte
	value := make(map[string][]byte, 1)
	value["secret"] = b

	return value, err
}

// SetupWithManager registers the controller with that manager so that it starts when the manager starts
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}
