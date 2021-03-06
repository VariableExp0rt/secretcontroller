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

const (
	namespaces                   = "default"
	secretType corev1.SecretType = "Opaque"
	myLabel                      = "mysupersecret"
)

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get;update;patch

// Reconcile is a reconciler for the core/v1 type Secret
func (r *SecretReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("secret", req.NamespacedName)

	//TODO: a check is needed here for the type of secret being reconciled

	//TODO
	//A switch statement is needed here to account the different secret types, which will determine
	//which functions are executed by 'case' ("generic", "tls", "docker-registry").

	var secret corev1.Secret
	if err := r.Get(ctx, req.NamespacedName, &secret); err != nil {
		log.Error(err, "unable to get secrets", "secret", secret)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if secret.Labels["secret-type"] == myLabel {
		log.Info("secret with appropriate label identified", "secret", secret)
	}

	_, notValid := r.compareTime(secret)
	fmt.Printf("%v-%T\t", notValid, notValid)
	// filter on the secrets with the labels where we know data is of a certain format -
	// for instance, we know that the secrets I've created are string, but more complex controllers
	// might store reconcile/manage certs that are stored in secrets

	patched, err := r.patchSecret(secret)
	if err != nil {
		log.Error(err, "unable to patch secret with new value", "secret", patched)
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		Requeue:      false,
		RequeueAfter: 0,
	}, nil

}

// compareTime is a function that evaluates whether a secret is more than 7 days old
// in such cases the secret that is more than 7 days old is returned to be used with secretGenerator
func (r *SecretReconciler) compareTime(s corev1.Secret) (corev1.Secret, float64) {

	// Will not work if the value is CreationTimestamp.Time, which returns null, more logic needed to make
	// everything UTC - works with straight UTC time though https://play.golang.org/p/vLi6bGIBj7d

	secretTime := s.CreationTimestamp.Time
	targetTime := time.Since(secretTime)

	notValid := targetTime.Hours()

	return s, notValid
}

func (r *SecretReconciler) filterSecret(secret corev1.Secret) (corev1.Secret, error) {
	secret, notValid := r.compareTime(secret)
	var err error

	if int(notValid) >= 168 {
		log.Info("expired secret identified")
	} else if int(notValid) < 168 {
		err = fmt.Errorf("ignoring secret because it has not expired: %v", secret.Name)
	}
	return secret, err
}

// secretGenerator is a function that will eventually take an int (which is the length of the secret to be generated)
// This will then be used within patchSecret
func (r *SecretReconciler) secretGenerator(s corev1.Secret) (corev1.Secret, map[string][]byte, error) {
	filtered, err := r.filterSecret(s)
	if err != nil {
		log.Infof("Skipped: %v", err)
	}

	//logic for generating a random string which is used as the secret value
	// first get the old data to be passed to the generateRandomBytes function
	old := s.Data

	value, err := r.generateRandomBytes(old)
	if err != nil {
		log.Errorf("Error creating new secret value for secret: %v", filtered)
	}

	return filtered, value, err
}

func (r *SecretReconciler) patchSecret(s corev1.Secret) (corev1.Secret, error) {
	ctx := context.TODO()
	secretToPatch, newSecretVal, err := r.secretGenerator(s)

	//patch logic to be generated for patching the object
	patch := s.DeepCopy()
	patch.Data = newSecretVal
	r.Patch(ctx, &secretToPatch, client.MergeFrom(patch))
	if err != nil {
		log.Errorf("unable to complete patch operation for secret", "secret", secretToPatch)
	}

	return secretToPatch, err
}

func (r *SecretReconciler) generateRandomBytes(oldval map[string][]byte) (map[string][]byte, error) {

	var newval map[string][]byte
	for key, val := range oldval {
		n := len(val)
		b := make([]byte, n)
		_, err := rand.Read(b)
		if err != nil {
			log.Error("unable to generate new value for secret data")
		}

		// Preserve the key, but assign new value to []byte and make newvalue the oldvalue
		// return the newvalue and use that in the patchSecret function to make the change
		// after deep copying the object
		oldval[key] = b
		newval = oldval
	}

	//TODO
	//logic to map []byte to map[string][]byte, but need to find a way of merging the patch such
	//that it retains the value of the string 'key'. "secret" is obviously not the original value
	return newval, nil
}

//Functions needed to handle other secret types, other than generic as above.

//TODO: account for StringData as well

// SetupWithManager registers the controller with that manager so that it starts when the manager starts
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}
