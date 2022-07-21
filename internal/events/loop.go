package events

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/runtime"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
)

// EventLoopConfig holds configuration parameters for EventLoop.
type EventLoopConfig struct {
	// Processor is the state ChangeProcessor.
	Processor state.ChangeProcessor
	// ServiceStore is the state ServiceStore.
	ServiceStore state.ServiceStore
	// SecretStore is the state SecretStore.
	SecretStore state.SecretStore
	// SecretMemoryManager is the state SecretMemoryManager.
	SecretMemoryManager state.SecretDiskMemoryManager
	// Generator is the nginx config Generator.
	Generator config.Generator
	// EventCh is a read-only channel for events.
	EventCh <-chan interface{}
	// Logger is the logger to be used by the EventLoop.
	Logger logr.Logger
	// NginxFileMgr is the file Manager for nginx.
	NginxFileMgr file.Manager
	// NginxRuntimeMgr manages nginx runtime.
	NginxRuntimeMgr runtime.Manager
	// StatusUpdater updates statuses on Kubernetes resources.
	StatusUpdater status.Updater
}

// EventLoop is the main event loop of the Gateway.
type EventLoop struct {
	cfg EventLoopConfig
}

// NewEventLoop creates a new EventLoop.
func NewEventLoop(cfg EventLoopConfig) *EventLoop {
	return &EventLoop{cfg: cfg}
}

// Start starts the EventLoop.
// The method will block until the EventLoop stops:
// - if it stops because of an error, the Start will return the error.
// - if it stops normally, the Start will return nil.
func (el *EventLoop) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			// although we always return nil, Start must return it to satisfy
			// "sigs.k8s.io/controller-runtime/pkg/manager".Runnable
			return nil
		case e := <-el.cfg.EventCh:
			el.handleEvent(ctx, e)
		}
	}
}

// FIXME(pleshakov): think about how to avoid using an interface{} here
func (el *EventLoop) handleEvent(ctx context.Context, event interface{}) {
	switch e := event.(type) {
	case *UpsertEvent:
		el.propagateUpsert(e)
	case *DeleteEvent:
		el.propagateDelete(e)
	default:
		panic(fmt.Errorf("unknown event type %T", e))
	}

	changed, conf, statuses := el.cfg.Processor.Process()
	if !changed {
		return
	}

	err := el.updateNginx(ctx, conf)
	if err != nil {
		el.cfg.Logger.Error(err, "Failed to update NGINX configuration")
	}

	el.cfg.StatusUpdater.Update(ctx, statuses)
}

func (el *EventLoop) updateNginx(ctx context.Context, conf state.Configuration) error {
	// Write all secrets (nuke and pave).
	// This will remove all secrets in the secrets directory before writing the requested secrets.
	// FIXME(kate-osborn): We may want to rethink this approach in the future and write and remove secrets individually.
	err := el.cfg.SecretMemoryManager.WriteAllRequestedSecrets()
	if err != nil {
		return err
	}

	cfg, warnings := el.cfg.Generator.Generate(conf)

	// For now, we keep all http servers in one config
	// We might rethink that. For example, we can write each server to its file
	// or group servers in some way.
	err = el.cfg.NginxFileMgr.WriteHTTPServersConfig("http-servers", cfg)
	if err != nil {
		return err
	}

	for obj, objWarnings := range warnings {
		for _, w := range objWarnings {
			// FIXME(pleshakov): report warnings via Object status
			el.cfg.Logger.Info("got warning while generating config",
				"kind", obj.GetObjectKind().GroupVersionKind().Kind,
				"namespace", obj.GetNamespace(),
				"name", obj.GetName(),
				"warning", w)
		}
	}

	return el.cfg.NginxRuntimeMgr.Reload(ctx)
}

func (el *EventLoop) propagateUpsert(e *UpsertEvent) {
	switch r := e.Resource.(type) {
	case *v1alpha2.GatewayClass:
		el.cfg.Processor.CaptureUpsertChange(r)
	case *v1alpha2.Gateway:
		el.cfg.Processor.CaptureUpsertChange(r)
	case *v1alpha2.HTTPRoute:
		el.cfg.Processor.CaptureUpsertChange(r)
	case *apiv1.Service:
		// FIXME(pleshakov): make sure the affected hosts are updated
		el.cfg.ServiceStore.Upsert(r)
	case *apiv1.Secret:
		// FIXME(kate-osborn): need to handle certificate rotation
		el.cfg.SecretStore.Upsert(r)
	default:
		panic(fmt.Errorf("unknown resource type %T", e.Resource))
	}
}

func (el *EventLoop) propagateDelete(e *DeleteEvent) {
	switch e.Type.(type) {
	case *v1alpha2.GatewayClass:
		el.cfg.Processor.CaptureDeleteChange(e.Type, e.NamespacedName)
	case *v1alpha2.Gateway:
		el.cfg.Processor.CaptureDeleteChange(e.Type, e.NamespacedName)
	case *v1alpha2.HTTPRoute:
		el.cfg.Processor.CaptureDeleteChange(e.Type, e.NamespacedName)
	case *apiv1.Service:
		// FIXME(pleshakov): make sure the affected hosts are updated
		el.cfg.ServiceStore.Delete(e.NamespacedName)
	case *apiv1.Secret:
		// FIXME(kate-osborn): make sure that affected servers are updated
		el.cfg.SecretStore.Delete(e.NamespacedName)
	default:
		panic(fmt.Errorf("unknown resource type %T", e.Type))
	}
}
