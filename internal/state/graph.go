package state

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// gateway represents the winning Gateway resource.
type gateway struct {
	// Source is the corresponding Gateway resource.
	Source *v1alpha2.Gateway
	// Listeners include the listeners of the Gateway.
	Listeners map[string]*listener
}

// route represents an HTTPRoute.
type route struct {
	// Source is the source resource of the route.
	// FIXME(pleshakov)
	// For now, we assume that the source is only HTTPRoute. Later we can support more types - TLSRoute, TCPRoute and UDPRoute.
	Source *v1alpha2.HTTPRoute

	// ValidSectionNameRefs includes the sectionNames from the parentRefs of the HTTPRoute that are valid -- i.e.
	// the Gateway resource has a corresponding valid listener.
	ValidSectionNameRefs map[string]struct{}
	// ValidSectionNameRefs includes the sectionNames from the parentRefs of the HTTPRoute that are invalid.
	InvalidSectionNameRefs map[string]struct{}
}

// gatewayClass represents the GatewayClass resource.
type gatewayClass struct {
	// Source is the source resource.
	Source *v1alpha2.GatewayClass
	// Valid shows whether the GatewayClass is valid.
	Valid bool
	// ErrorMsg explains the error when the resource is not valid.
	ErrorMsg string
}

// graph is a graph-like representation of Gateway API resources.
type graph struct {
	// GatewayClass holds the GatewayClass resource.
	GatewayClass *gatewayClass
	// Gateway holds the winning Gateway resource.
	Gateway *gateway
	// IgnoredGateways holds the ignored Gateway resources, which belong to the NGINX Gateway (based on the
	// GatewayClassName field of the resource) but ignored. It doesn't hold the Gateway resources that do not belong to
	// the NGINX Gateway.
	IgnoredGateways map[types.NamespacedName]*v1alpha2.Gateway
	// Routes holds route resources.
	Routes map[types.NamespacedName]*route
}

// buildGraph builds a graph from a store assuming that the Gateway resource has the gwNsName namespace and name.
func buildGraph(
	store *store,
	controllerName string,
	gcName string,
	secretMemoryMgr SecretDiskMemoryManager,
) *graph {
	gc := buildGatewayClass(store.gc, controllerName)

	gw, ignoredGws := processGateways(store.gateways, gcName)

	listeners := buildListeners(gw, gcName, secretMemoryMgr)

	routes := make(map[types.NamespacedName]*route)
	for _, ghr := range store.httpRoutes {
		ignored, r := bindHTTPRouteToListeners(ghr, gw, ignoredGws, listeners)
		if !ignored {
			routes[getNamespacedName(ghr)] = r
		}
	}

	g := &graph{
		GatewayClass:    gc,
		Routes:          routes,
		IgnoredGateways: ignoredGws,
	}

	if gw != nil {
		g.Gateway = &gateway{
			Source:    gw,
			Listeners: listeners,
		}
	}

	return g
}

// processGateways determines which Gateway resource the NGINX Gateway will use (the winner) and which Gateway(s) will
// be ignored. Note that the function will not take into the account any unrelated Gateway resources - the ones with the
// different GatewayClassName field.
func processGateways(gws map[types.NamespacedName]*v1alpha2.Gateway, gcName string) (winner *v1alpha2.Gateway, ignoredGateways map[types.NamespacedName]*v1alpha2.Gateway) {
	referencedGws := make([]*v1alpha2.Gateway, 0, len(gws))

	for _, gw := range gws {
		if string(gw.Spec.GatewayClassName) != gcName {
			continue
		}

		referencedGws = append(referencedGws, gw)
	}

	if len(referencedGws) == 0 {
		return nil, nil
	}

	sort.Slice(referencedGws, func(i, j int) bool {
		return lessObjectMeta(&referencedGws[i].ObjectMeta, &referencedGws[j].ObjectMeta)
	})

	ignoredGws := make(map[types.NamespacedName]*v1alpha2.Gateway)

	for _, gw := range referencedGws[1:] {
		ignoredGws[getNamespacedName(gw)] = gw
	}

	return referencedGws[0], ignoredGws
}

func buildGatewayClass(gc *v1alpha2.GatewayClass, controllerName string) *gatewayClass {
	if gc == nil {
		return nil
	}

	var errorMsg string

	err := validateGatewayClass(gc, controllerName)
	if err != nil {
		errorMsg = err.Error()
	}

	return &gatewayClass{
		Source:   gc,
		Valid:    err == nil,
		ErrorMsg: errorMsg,
	}
}

func buildListeners(gw *v1alpha2.Gateway, gcName string, secretMemoryMgr SecretDiskMemoryManager) map[string]*listener {
	listeners := make(map[string]*listener)

	if gw == nil || string(gw.Spec.GatewayClassName) != gcName {
		return listeners
	}

	listenerFactory := newListenerConfiguratorFactory(gw, secretMemoryMgr)

	for _, gl := range gw.Spec.Listeners {
		configurator := listenerFactory.getConfiguratorForListener(gl)
		listeners[string(gl.Name)] = configurator.configure(gl)
	}

	return listeners
}

// bindHTTPRouteToListeners tries to bind an HTTPRoute to listener.
// There are three possibilities:
// (1) HTTPRoute will be ignored.
// (2) HTTPRoute will be processed but not bound.
// (3) HTTPRoute will be processed and bound to a listener.
func bindHTTPRouteToListeners(
	ghr *v1alpha2.HTTPRoute,
	gw *v1alpha2.Gateway,
	ignoredGws map[types.NamespacedName]*v1alpha2.Gateway,
	listeners map[string]*listener,
) (ignored bool, r *route) {
	if len(ghr.Spec.ParentRefs) == 0 {
		// ignore HTTPRoute without refs
		return true, nil
	}

	r = &route{
		Source:                 ghr,
		ValidSectionNameRefs:   make(map[string]struct{}),
		InvalidSectionNameRefs: make(map[string]struct{}),
	}

	// FIXME (pleshakov) Handle the case when parent refs are duplicated

	processed := false

	for _, p := range ghr.Spec.ParentRefs {
		// FIXME(pleshakov) Support empty section name
		if p.SectionName == nil || *p.SectionName == "" {
			continue
		}

		// if the namespace is missing, assume the namespace of the HTTPRoute
		ns := ghr.Namespace
		if p.Namespace != nil {
			ns = string(*p.Namespace)
		}

		name := string(*p.SectionName)

		// Below we will figure out what Gateway resource the parentRef references and act accordingly. There are 3 cases.

		// Case 1: the parentRef references the winning Gateway.

		if gw != nil && gw.Namespace == ns && gw.Name == string(p.Name) {

			// Find a listener

			// FIXME(pleshakov)
			// For now, let's do simple matching.
			// However, we need to also support wildcard matching.
			// More over, we need to handle cases when a Route host matches multiple HTTP listeners on the same port when
			// sectionName is empty and only choose one listener.
			// For example:
			// - Route with host foo.example.com;
			// - listener 1 for port 80 with hostname foo.example.com
			// - listener 2 for port 80 with hostname *.example.com;
			// In this case, the Route host foo.example.com should choose listener 1, as it is a more specific match.

			processed = true

			l, exists := listeners[name]
			if !exists {
				r.InvalidSectionNameRefs[name] = struct{}{}
				continue
			}

			accepted := findAcceptedHostnames(l.Source.Hostname, ghr.Spec.Hostnames)

			if len(accepted) > 0 {
				for _, h := range accepted {
					l.AcceptedHostnames[h] = struct{}{}
				}
				r.ValidSectionNameRefs[name] = struct{}{}
				l.Routes[getNamespacedName(ghr)] = r
			} else {
				r.InvalidSectionNameRefs[name] = struct{}{}
			}

			continue
		}

		// Case 2: the parentRef references an ignored Gateway resource.

		key := types.NamespacedName{Namespace: ns, Name: string(p.Name)}

		if _, exist := ignoredGws[key]; exist {
			r.InvalidSectionNameRefs[name] = struct{}{}

			processed = true
			continue
		}

		// Case 3: the parentRef references some unrelated to this NGINX Gateway Gateway or other resource.

		// Do nothing
	}

	if !processed {
		return true, nil
	}

	return false, r
}

func findAcceptedHostnames(listenerHostname *v1alpha2.Hostname, routeHostnames []v1alpha2.Hostname) []string {
	hostname := getHostname(listenerHostname)

	match := func(h v1alpha2.Hostname) bool {
		if hostname == "" {
			return true
		}
		return string(h) == hostname
	}

	var result []string

	for _, h := range routeHostnames {
		if match(h) {
			result = append(result, string(h))
		}
	}

	return result
}

func getHostname(h *v1alpha2.Hostname) string {
	if h == nil {
		return ""
	}
	return string(*h)
}

func validateGatewayClass(gc *v1alpha2.GatewayClass, controllerName string) error {
	if string(gc.Spec.ControllerName) != controllerName {
		return fmt.Errorf("Spec.ControllerName must be %s got %s", controllerName, gc.Spec.ControllerName)
	}

	return nil
}
