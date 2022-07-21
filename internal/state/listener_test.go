package state

import (
	"testing"

	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func TestValidateHTTPListener(t *testing.T) {
	tests := []struct {
		l        v1alpha2.Listener
		expected bool
		msg      string
	}{
		{
			l: v1alpha2.Listener{
				Port:     80,
				Protocol: v1alpha2.HTTPProtocolType,
			},
			expected: true,
			msg:      "valid",
		},
		{
			l: v1alpha2.Listener{
				Port:     81,
				Protocol: v1alpha2.HTTPProtocolType,
			},
			expected: false,
			msg:      "invalid port",
		},
	}

	for _, test := range tests {
		result := validateHTTPListener(test.l)
		if result != test.expected {
			t.Errorf("validateListener() returned %v but expected %v for the case of %q", result, test.expected, test.msg)
		}
	}
}

func TestValidateHTTPSListener(t *testing.T) {
	gwNs := "gateway-ns"

	validSecretRef := &v1alpha2.SecretObjectReference{
		Kind:      (*v1alpha2.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer(gwNs)),
	}

	invalidSecretRefType := &v1alpha2.SecretObjectReference{
		Kind:      (*v1alpha2.Kind)(helpers.GetStringPointer("ConfigMap")),
		Name:      "secret",
		Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer(gwNs)),
	}

	invalidSecretRefTNamespace := &v1alpha2.SecretObjectReference{
		Kind:      (*v1alpha2.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer("diff-ns")),
	}

	tests := []struct {
		l        v1alpha2.Listener
		expected bool
		msg      string
	}{
		{
			l: v1alpha2.Listener{
				Port:     443,
				Protocol: v1alpha2.HTTPSProtocolType,
				TLS: &v1alpha2.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1alpha2.TLSModeTerminate),
					CertificateRefs: []*v1alpha2.SecretObjectReference{validSecretRef},
				},
			},
			expected: true,
			msg:      "valid",
		},
		{
			l: v1alpha2.Listener{
				Port:     80,
				Protocol: v1alpha2.HTTPSProtocolType,
				TLS: &v1alpha2.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1alpha2.TLSModeTerminate),
					CertificateRefs: []*v1alpha2.SecretObjectReference{validSecretRef},
				},
			},
			expected: false,
			msg:      "invalid port",
		},
		{
			l: v1alpha2.Listener{
				Port:     443,
				Protocol: v1alpha2.HTTPSProtocolType,
				TLS: &v1alpha2.GatewayTLSConfig{
					Mode: helpers.GetTLSModePointer(v1alpha2.TLSModeTerminate),
				},
			},
			expected: false,
			msg:      "invalid - no cert ref",
		},
		{
			l: v1alpha2.Listener{
				Port:     443,
				Protocol: v1alpha2.HTTPSProtocolType,
				TLS: &v1alpha2.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1alpha2.TLSModePassthrough),
					CertificateRefs: []*v1alpha2.SecretObjectReference{validSecretRef},
				},
			},
			expected: false,
			msg:      "invalid tls mode",
		},
		{
			l: v1alpha2.Listener{
				Port:     443,
				Protocol: v1alpha2.HTTPSProtocolType,
				TLS: &v1alpha2.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1alpha2.TLSModeTerminate),
					CertificateRefs: []*v1alpha2.SecretObjectReference{invalidSecretRefType},
				},
			},
			expected: false,
			msg:      "invalid cert ref kind",
		},
		{
			l: v1alpha2.Listener{
				Port:     443,
				Protocol: v1alpha2.HTTPSProtocolType,
				TLS: &v1alpha2.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1alpha2.TLSModeTerminate),
					CertificateRefs: []*v1alpha2.SecretObjectReference{invalidSecretRefTNamespace},
				},
			},
			expected: false,
			msg:      "invalid cert ref namespace",
		},
		{
			l: v1alpha2.Listener{
				Port:     443,
				Protocol: v1alpha2.HTTPSProtocolType,
			},
			expected: false,
			msg:      "invalid - no tls config",
		},
	}

	for _, test := range tests {
		result := validateHTTPSListener(test.l, gwNs)
		if result != test.expected {
			t.Errorf("validateHTTPSListener() returned %v but expected %v for the case of %q", result, test.expected, test.msg)
		}
	}
}
