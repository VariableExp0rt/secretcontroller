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
	"fmt"
	"time"

	"github.com/go-logr/logr"
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
		log.Error(err, "unable to get secrets", "secrets:", secret)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if _, err := r.compareTime(secret); err != nil {
		log.Error(err, "unable to compare creation timestamp time for secrets")
		return ctrl.Result{}, err
	}

	if _, err := r.secretGenerator(8); err != nil {
		log.Error(err, "unable to generate new secret")
		return ctrl.Result{}, err
	}

	// if err := patchSecret(secret, newSecret); err != nil {
	//log.Error(err, "unable to patch secret with new value", "new_value", newSecret)
	//}

	//move patching to the function below to avoid issues
	//if err := r.Patch(ctx, secrets.DeepCopyObject(), client.RawPatch(types.JSONPatchType, newSecret), &client.PatchOptions{}) {
	//	log.Info("reconiling and updating secret object with new secret value")
	//}

	return ctrl.Result{}, nil

}

// compareTime is a function that evaluates whether a secret is more than 7 days old
// in such cases the secret that is more than 7 days old is returned to be used with secretGenerator
func (r *SecretReconciler) compareTime(secrets corev1.Secret) (secret corev1.Secret, err error) {
	secretTime := secret.CreationTimestamp.Time
	targetTime := time.Now().AddDate(0, 0, -7)

	isNotValid := secretTime.Before(targetTime)
	if isNotValid == true {
		fmt.Printf("secret: %v is older than 7 days (%v), forbidden for this applciation", secret, secretTime)
	}
	return secret, err
}

// secretGenerator is a function that will eventually take an int (which is the length of the secret to be generated)
// This will then be used within patchSecret
func (r *SecretReconciler) secretGenerator(length int) (newSecret map[string][]byte, err error) {
	newSecret = make(map[string][]byte, 1)

	return newSecret, err
}

func (r *SecretReconciler) patchSecret(secret, newSecret map[string][]byte) error {

	var err error
	return err
}

// SetupWithManager registers the controller with that manager so that it starts when the manager starts
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}
