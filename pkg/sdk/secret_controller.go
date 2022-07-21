package sdk

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type secretReconciler struct {
	client.Client
	scheme *runtime.Scheme
	impl   SecretImpl
}

// RegisterSecretController registers the SecretController in the manager.
func RegisterSecretController(mgr manager.Manager, impl SecretImpl) error {
	r := &secretReconciler{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		impl:   impl,
	}

	return ctlr.NewControllerManagedBy(mgr).
		For(&apiv1.Secret{}).
		Complete(r)
}

func (r *secretReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithValues("secret", req.NamespacedName)

	log.V(3).Info("Reconciling Secret")

	found := true
	var secret apiv1.Secret
	err := r.Get(ctx, req.NamespacedName, &secret)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to get Secret")
			return reconcile.Result{}, err
		}
		found = false
	}

	if !found {
		log.V(3).Info("Removing Secret")

		r.impl.Remove(req.NamespacedName)
		return reconcile.Result{}, nil
	}

	log.V(3).Info("Upserting Secret")

	r.impl.Upsert(&secret)
	return reconcile.Result{}, nil
}
