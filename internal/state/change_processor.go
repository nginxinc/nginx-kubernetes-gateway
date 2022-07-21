package state

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ChangeProcessor

// ChangeProcessor processes the changes to resources producing the internal representation of the Gateway configuration.
// ChangeProcessor only supports one Gateway resource.
type ChangeProcessor interface {
	// CaptureUpsertChange captures an upsert change to a resource.
	// It panics if the resource is of unsupported type or if the passed Gateway is different from the one this ChangeProcessor
	// was created for.
	CaptureUpsertChange(obj client.Object)
	// CaptureDeleteChange captures a delete change to a resource.
	// The method panics if the resource is of unsupported type or if the passed Gateway is different from the one this ChangeProcessor
	// was created for.
	CaptureDeleteChange(resourceType client.Object, nsname types.NamespacedName)
	// Process processes any captured changes and produces an internal representation of the Gateway configuration and
	// the status information about the processed resources.
	// If no changes were captured, the changed return argument will be false and both the configuration and statuses
	// will be empty.
	Process() (changed bool, conf Configuration, statuses Statuses)
}

// ChangeProcessorConfig holds configuration parameters for ChangeProcessorImpl.
type ChangeProcessorConfig struct {
	// GatewayNsName is the namespaced name of the Gateway resource.
	GatewayNsName types.NamespacedName
	// GatewayCtlrName is the name of the Gateway controller.
	GatewayCtlrName string
	// GatewayClassName is the name of the GatewayClass resource.
	GatewayClassName string
	// SecretMemoryManager is the secret memory manager.
	SecretMemoryManager SecretDiskMemoryManager
}

type ChangeProcessorImpl struct {
	store   *store
	changed bool
	cfg     ChangeProcessorConfig

	lock sync.Mutex
}

// NewChangeProcessorImpl creates a new ChangeProcessorImpl for the Gateway resource with the configured namespace name.
func NewChangeProcessorImpl(cfg ChangeProcessorConfig) *ChangeProcessorImpl {
	return &ChangeProcessorImpl{
		store: newStore(),
		cfg:   cfg,
	}
}

// FIXME(pleshakov)
// Currently, changes (upserts/delete) trigger rebuilding of the configuration, even if the change doesn't change
// the configuration or the statuses of the resources. For example, a change in a Gateway resource that doesn't
// belong to the NGINX Gateway or an HTTPRoute that doesn't belong to any of the Gateways of the NGINX Gateway.
// Find a way to ignore changes that don't affect the configuration and/or statuses of the resources.

func (c *ChangeProcessorImpl) CaptureUpsertChange(obj client.Object) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.changed = true

	switch o := obj.(type) {
	case *v1alpha2.GatewayClass:
		if o.Name != c.cfg.GatewayClassName {
			panic(fmt.Errorf("gatewayclass resource must be %s, got %s", c.cfg.GatewayClassName, o.Name))
		}
		// if the resource spec hasn't changed (its generation is the same), ignore the upsert
		if c.store.gc != nil && c.store.gc.Generation == o.Generation {
			c.changed = false
		}
		c.store.gc = o
	case *v1alpha2.Gateway:
		// if the resource spec hasn't changed (its generation is the same), ignore the upsert
		prev, exist := c.store.gateways[getNamespacedName(obj)]
		if exist && o.Generation == prev.Generation {
			c.changed = false
		}
		c.store.gateways[getNamespacedName(obj)] = o
	case *v1alpha2.HTTPRoute:
		// if the resource spec hasn't changed (its generation is the same), ignore the upsert
		prev, exist := c.store.httpRoutes[getNamespacedName(obj)]
		if exist && o.Generation == prev.Generation {
			c.changed = false
		}
		c.store.httpRoutes[getNamespacedName(obj)] = o
	default:
		panic(fmt.Errorf("ChangeProcessor doesn't support %T", obj))
	}
}

func (c *ChangeProcessorImpl) CaptureDeleteChange(resourceType client.Object, nsname types.NamespacedName) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.changed = true

	switch resourceType.(type) {
	case *v1alpha2.GatewayClass:
		if nsname.Name != c.cfg.GatewayClassName {
			panic(fmt.Errorf("gatewayclass resource must be %s, got %s", c.cfg.GatewayClassName, nsname.Name))
		}
		c.store.gc = nil
	case *v1alpha2.Gateway:
		delete(c.store.gateways, nsname)
	case *v1alpha2.HTTPRoute:
		delete(c.store.httpRoutes, nsname)
	default:
		panic(fmt.Errorf("ChangeProcessor doesn't support %T", resourceType))
	}
}

func (c *ChangeProcessorImpl) Process() (changed bool, conf Configuration, statuses Statuses) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.changed {
		return false, conf, statuses
	}

	c.changed = false

	graph := buildGraph(
		c.store,
		c.cfg.GatewayCtlrName,
		c.cfg.GatewayClassName,
		c.cfg.SecretMemoryManager,
	)

	conf = buildConfiguration(graph)
	statuses = buildStatuses(graph)

	return true, conf, statuses
}
