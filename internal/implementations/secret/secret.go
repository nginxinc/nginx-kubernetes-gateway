package implementation

import (
	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

type secretImplementation struct {
	conf    config.Config
	eventCh chan<- interface{}
}

// NewSecretImplementation creates a new SecretImplementation.
func NewSecretImplementation(cfg config.Config, eventCh chan<- interface{}) sdk.SecretImpl {
	return &secretImplementation{
		conf:    cfg,
		eventCh: eventCh,
	}
}

func (impl *secretImplementation) Logger() logr.Logger {
	return impl.conf.Logger
}

func (impl secretImplementation) Upsert(secret *apiv1.Secret) {
	impl.Logger().Info(
		"Secret was upserted",
		"namespace", secret.Namespace,
		"name", secret.Name,
	)

	impl.eventCh <- &events.UpsertEvent{
		Resource: secret,
	}
}

func (impl secretImplementation) Remove(nsname types.NamespacedName) {
	impl.Logger().Info(
		"Secret was removed",
		"namespace", nsname.Namespace,
		"name", nsname.Name,
	)

	impl.eventCh <- &events.DeleteEvent{
		NamespacedName: nsname,
		Type:           &apiv1.Secret{},
	}
}
