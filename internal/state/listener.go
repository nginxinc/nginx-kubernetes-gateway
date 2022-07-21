package state

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// listener represents a listener of the Gateway resource.
// FIXME(pleshakov) For now, we only support HTTP and HTTPS listeners.
type listener struct {
	// Source holds the source of the listener from the Gateway resource.
	Source v1alpha2.Listener
	// Valid shows whether the listener is valid.
	Valid bool
	// SecretPath is the path to the secret on disk.
	SecretPath string
	// Routes holds the routes attached to the listener.
	Routes map[types.NamespacedName]*route
	// AcceptedHostnames is an intersection between the hostnames supported by the listener and the hostnames
	// from the attached routes.
	AcceptedHostnames map[string]struct{}
}

type listenerConfigurator interface {
	configure(listener v1alpha2.Listener) *listener
}

type listenerConfiguratorFactory struct {
	https *httpsListenerConfigurator
	http  *httpListenerConfigurator
}

func (f *listenerConfiguratorFactory) getConfiguratorForListener(l v1alpha2.Listener) listenerConfigurator {
	switch l.Protocol {
	case v1alpha2.HTTPProtocolType:
		return f.http
	case v1alpha2.HTTPSProtocolType:
		return f.https
	default:
		return newInvalidProtocolListenerConfigurator()
	}
}

func newListenerConfiguratorFactory(gw *v1alpha2.Gateway, secretMemoryMgr SecretDiskMemoryManager) *listenerConfiguratorFactory {
	return &listenerConfiguratorFactory{
		https: newHTTPSListenerConfigurator(gw, secretMemoryMgr),
		http:  newHTTPListenerConfigurator(),
	}
}

type httpsListenerConfigurator struct {
	gateway         *v1alpha2.Gateway
	secretMemoryMgr SecretDiskMemoryManager
	usedHostnames   map[string]*listener
}

func newHTTPSListenerConfigurator(gateway *v1alpha2.Gateway, secretMemoryMgr SecretDiskMemoryManager) *httpsListenerConfigurator {
	return &httpsListenerConfigurator{
		gateway:         gateway,
		secretMemoryMgr: secretMemoryMgr,
		usedHostnames:   make(map[string]*listener),
	}
}

func (c *httpsListenerConfigurator) configure(gl v1alpha2.Listener) *listener {
	var path string
	var err error

	valid := validateHTTPSListener(gl, c.gateway.Namespace)

	if valid {
		nsname := types.NamespacedName{
			Namespace: c.gateway.Namespace,
			Name:      string(gl.TLS.CertificateRefs[0].Name),
		}

		path, err = c.secretMemoryMgr.Request(nsname)
		if err != nil {
			valid = false
		}
	}

	h := getHostname(gl.Hostname)

	if holder, exist := c.usedHostnames[h]; exist {
		valid = false
		holder.Valid = false // all listeners for the same hostname become conflicted
	}

	l := &listener{
		Source:            gl,
		Valid:             valid,
		SecretPath:        path,
		Routes:            make(map[types.NamespacedName]*route),
		AcceptedHostnames: make(map[string]struct{}),
	}

	c.usedHostnames[h] = l

	return l
}

type httpListenerConfigurator struct {
	usedHostnames map[string]*listener
}

func newHTTPListenerConfigurator() *httpListenerConfigurator {
	return &httpListenerConfigurator{
		usedHostnames: make(map[string]*listener),
	}
}

func (c *httpListenerConfigurator) configure(gl v1alpha2.Listener) *listener {
	valid := validateHTTPListener(gl)

	h := getHostname(gl.Hostname)

	if holder, exist := c.usedHostnames[h]; exist {
		valid = false
		holder.Valid = false // all listeners for the same hostname become conflicted
	}

	l := &listener{
		Source:            gl,
		Valid:             valid,
		Routes:            make(map[types.NamespacedName]*route),
		AcceptedHostnames: make(map[string]struct{}),
	}

	c.usedHostnames[h] = l

	return l
}

type invalidProtocolListenerConfigurator struct{}

func newInvalidProtocolListenerConfigurator() *invalidProtocolListenerConfigurator {
	return &invalidProtocolListenerConfigurator{}
}

func (c *invalidProtocolListenerConfigurator) configure(gl v1alpha2.Listener) *listener {
	return &listener{
		Source:            gl,
		Valid:             false,
		Routes:            make(map[types.NamespacedName]*route),
		AcceptedHostnames: make(map[string]struct{}),
	}
}

func validateHTTPListener(listener v1alpha2.Listener) bool {
	// FIXME(pleshakov): For now we require that all HTTP listeners bind to port 80
	return listener.Port == 80
}

func validateHTTPSListener(listener v1alpha2.Listener, gwNsname string) bool {
	// FIXME(kate-osborn):
	// 1. For now we require that all HTTPS listeners bind to port 443
	// 2. Only TLSModeTerminate is supported.
	if listener.Port != 443 || listener.TLS == nil || *listener.TLS.Mode != v1alpha2.TLSModeTerminate || len(listener.TLS.CertificateRefs) == 0 {
		return false
	}

	certRef := listener.TLS.CertificateRefs[0]
	// certRef Kind has default of "Secret" so it's safe to directly access the Kind here
	if *certRef.Kind != "Secret" {
		return false
	}

	// secret must be in the same namespace as the gateway
	if certRef.Namespace != nil && string(*certRef.Namespace) != gwNsname {
		return false
	}

	return true
}
