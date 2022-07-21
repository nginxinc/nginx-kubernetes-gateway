package implementation

import (
	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

type serviceImplementation struct {
	conf    config.Config
	eventCh chan<- interface{}
}

// FIXME(pleshakov): serviceImplementation looks similar to httpRouteImplemenation
// consider if it is possible to reduce the amount of code.

// NewServiceImplementation creates a new ServiceImplementation.
func NewServiceImplementation(cfg config.Config, eventCh chan<- interface{}) sdk.ServiceImpl {
	return &serviceImplementation{
		conf:    cfg,
		eventCh: eventCh,
	}
}

func (impl *serviceImplementation) Logger() logr.Logger {
	return impl.conf.Logger
}

func (impl *serviceImplementation) Upsert(svc *apiv1.Service) {
	impl.Logger().Info("Service was upserted",
		"namespace", svc.Namespace, "name", svc.Name,
	)

	impl.eventCh <- &events.UpsertEvent{
		Resource: svc,
	}
}

func (impl *serviceImplementation) Remove(nsname types.NamespacedName) {
	impl.Logger().Info("Service resource was removed",
		"namespace", nsname.Namespace, "name", nsname.Name,
	)

	impl.eventCh <- &events.DeleteEvent{
		NamespacedName: nsname,
		Type:           &apiv1.Service{},
	}
}
