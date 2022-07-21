// nolint:gosec
package state

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

var testSecret = &v1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "secret",
		Namespace: "test",
	},
	Data: map[string][]byte{
		v1.TLSCertKey: []byte(`-----BEGIN CERTIFICATE-----
MIIDLjCCAhYCCQDAOF9tLsaXWjANBgkqhkiG9w0BAQsFADBaMQswCQYDVQQGEwJV
UzELMAkGA1UECAwCQ0ExITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0
ZDEbMBkGA1UEAwwSY2FmZS5leGFtcGxlLmNvbSAgMB4XDTE4MDkxMjE2MTUzNVoX
DTIzMDkxMTE2MTUzNVowWDELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMSEwHwYD
VQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxGTAXBgNVBAMMEGNhZmUuZXhh
bXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCp6Kn7sy81
p0juJ/cyk+vCAmlsfjtFM2muZNK0KtecqG2fjWQb55xQ1YFA2XOSwHAYvSdwI2jZ
ruW8qXXCL2rb4CZCFxwpVECrcxdjm3teViRXVsYImmJHPPSyQgpiobs9x7DlLc6I
BA0ZjUOyl0PqG9SJexMV73WIIa5rDVSF2r4kSkbAj4Dcj7LXeFlVXH2I5XwXCptC
n67JCg42f+k8wgzcRVp8XZkZWZVjwq9RUKDXmFB2YyN1XEWdZ0ewRuKYUJlsm692
skOrKQj0vkoPn41EE/+TaVEpqLTRoUY3rzg7DkdzfdBizFO2dsPNFx2CW0jXkNLv
Ko25CZrOhXAHAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAKHFCcyOjZvoHswUBMdL
RdHIb383pWFynZq/LuUovsVA58B0Cg7BEfy5vWVVrq5RIkv4lZ81N29x21d1JH6r
jSnQx+DXCO/TJEV5lSCUpIGzEUYaUPgRyjsM/NUdCJ8uHVhZJ+S6FA+CnOD9rn2i
ZBePCI5rHwEXwnnl8ywij3vvQ5zHIuyBglWr/Qyui9fjPpwWUvUm4nv5SMG9zCV7
PpuwvuatqjO1208BjfE/cZHIg8Hw9mvW9x9C+IQMIMDE7b/g6OcK7LGTLwlFxvA8
7WjEequnayIphMhKRXVf1N349eN98Ez38fOTHTPbdJjFA/PcC+Gyme+iGt5OQdFh
yRE=
-----END CERTIFICATE-----`),
		v1.TLSPrivateKeyKey: []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAqeip+7MvNadI7if3MpPrwgJpbH47RTNprmTStCrXnKhtn41k
G+ecUNWBQNlzksBwGL0ncCNo2a7lvKl1wi9q2+AmQhccKVRAq3MXY5t7XlYkV1bG
CJpiRzz0skIKYqG7Pcew5S3OiAQNGY1DspdD6hvUiXsTFe91iCGuaw1Uhdq+JEpG
wI+A3I+y13hZVVx9iOV8FwqbQp+uyQoONn/pPMIM3EVafF2ZGVmVY8KvUVCg15hQ
dmMjdVxFnWdHsEbimFCZbJuvdrJDqykI9L5KD5+NRBP/k2lRKai00aFGN684Ow5H
c33QYsxTtnbDzRcdgltI15DS7yqNuQmazoVwBwIDAQABAoIBAQCPSdSYnQtSPyql
FfVFpTOsoOYRhf8sI+ibFxIOuRauWehhJxdm5RORpAzmCLyL5VhjtJme223gLrw2
N99EjUKb/VOmZuDsBc6oCF6QNR58dz8cnORTewcotsJR1pn1hhlnR5HqJJBJask1
ZEnUQfcXZrL94lo9JH3E+Uqjo1FFs8xxE8woPBqjZsV7pRUZgC3LhxnwLSExyFo4
cxb9SOG5OmAJozStFoQ2GJOes8rJ5qfdvytgg9xbLaQL/x0kpQ62BoFMBDdqOePW
KfP5zZ6/07/vpj48yA1Q32PzobubsBLd3Kcn32jfm1E7prtWl+JeOFiOznBQFJbN
4qPVRz5hAoGBANtWyxhNCSLu4P+XgKyckljJ6F5668fNj5CzgFRqJ09zn0TlsNro
FTLZcxDqnR3HPYM42JERh2J/qDFZynRQo3cg3oeivUdBVGY8+FI1W0qdub/L9+yu
edOZTQ5XmGGp6r6jexymcJim/OsB3ZnYOpOrlD7SPmBvzNLk4MF6gxbXAoGBAMZO
0p6HbBmcP0tjFXfcKE77ImLm0sAG4uHoUx0ePj/2qrnTnOBBNE4MvgDuTJzy+caU
k8RqmdHCbHzTe6fzYq/9it8sZ77KVN1qkbIcuc+RTxA9nNh1TjsRne74Z0j1FCLk
hHcqH0ri7PYSKHTE8FvFCxZYdbuB84CmZihvxbpRAoGAIbjqaMYPTYuklCda5S79
YSFJ1JzZe1Kja//tDw1zFcgVCKa31jAwciz0f/lSRq3HS1GGGmezhPVTiqLfeZqc
R0iKbhgbOcVVkJJ3K0yAyKwPTumxKHZ6zImZS0c0am+RY9YGq5T7YrzpzcfvpiOU
ffe3RyFT7cfCmfoOhDCtzukCgYB30oLC1RLFOrqn43vCS51zc5zoY44uBzspwwYN
TwvP/ExWMf3VJrDjBCH+T/6sysePbJEImlzM+IwytFpANfiIXEt/48Xf60Nx8gWM
uHyxZZx/NKtDw0V8vX1POnq2A5eiKa+8jRARYKJLYNdfDuwolxvG6bZhkPi/4EtT
3Y18sQKBgHtKbk+7lNJVeswXE5cUG6EDUsDe/2Ua7fXp7FcjqBEoap1LSw+6TXp0
ZgrmKE8ARzM47+EJHUviiq/nupE15g0kJW3syhpU9zZLO7ltB0KIkO9ZRcmUjo8Q
cpLlHMAqbLJ8WYGJCkhiWxyal6hYTyWY4cVkC0xtTl/hUE9IeNKo
-----END RSA PRIVATE KEY-----`),
	},
	Type: v1.SecretTypeTLS,
}

var (
	secretPath       = "/etc/nginx/secrets/test_secret"
	secretsDirectory = "/etc/nginx/secrets"
)

func TestBuildGraph(t *testing.T) {
	const (
		gcName         = "my-class"
		controllerName = "my.controller"
	)

	createRoute := func(name string, gatewayName string, listenerName string) *v1alpha2.HTTPRoute {
		return &v1alpha2.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
			Spec: v1alpha2.HTTPRouteSpec{
				CommonRouteSpec: v1alpha2.CommonRouteSpec{
					ParentRefs: []v1alpha2.ParentRef{
						{
							Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
							Name:        v1alpha2.ObjectName(gatewayName),
							SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer(listenerName)),
						},
					},
				},
				Hostnames: []v1alpha2.Hostname{
					"foo.example.com",
				},
				Rules: []v1alpha2.HTTPRouteRule{
					{
						Matches: []v1alpha2.HTTPRouteMatch{
							{
								Path: &v1alpha2.HTTPPathMatch{
									Value: helpers.GetStringPointer("/"),
								},
							},
						},
					},
				},
			},
		}
	}

	hr1 := createRoute("hr-1", "gateway-1", "listener-80-1")
	hr2 := createRoute("hr-2", "wrong-gateway", "listener-80-1")
	hr3 := createRoute("hr-3", "gateway-1", "listener-443-1") // https listener; should not conflict with hr1

	createGateway := func(name string) *v1alpha2.Gateway {
		return &v1alpha2.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
			Spec: v1alpha2.GatewaySpec{
				GatewayClassName: gcName,
				Listeners: []v1alpha2.Listener{
					{
						Name:     "listener-80-1",
						Hostname: nil,
						Port:     80,
						Protocol: v1alpha2.HTTPProtocolType,
					},

					{
						Name:     "listener-443-1",
						Hostname: nil,
						Port:     443,
						TLS: &v1alpha2.GatewayTLSConfig{
							Mode: helpers.GetTLSModePointer(v1alpha2.TLSModeTerminate),
							CertificateRefs: []*v1alpha2.SecretObjectReference{
								{
									Kind:      (*v1alpha2.Kind)(helpers.GetStringPointer("Secret")),
									Name:      "secret",
									Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
								},
							},
						},
						Protocol: v1alpha2.HTTPSProtocolType,
					},
				},
			},
		}
	}

	gw1 := createGateway("gateway-1")
	gw2 := createGateway("gateway-2")

	store := &store{
		gc: &v1alpha2.GatewayClass{
			Spec: v1alpha2.GatewayClassSpec{
				ControllerName: controllerName,
			},
		},
		gateways: map[types.NamespacedName]*v1alpha2.Gateway{
			{Namespace: "test", Name: "gateway-1"}: gw1,
			{Namespace: "test", Name: "gateway-2"}: gw2,
		},
		httpRoutes: map[types.NamespacedName]*v1alpha2.HTTPRoute{
			{Namespace: "test", Name: "hr-1"}: hr1,
			{Namespace: "test", Name: "hr-2"}: hr2,
			{Namespace: "test", Name: "hr-3"}: hr3,
		},
	}

	routeHR1 := &route{
		Source: hr1,
		ValidSectionNameRefs: map[string]struct{}{
			"listener-80-1": {},
		},
		InvalidSectionNameRefs: map[string]struct{}{},
	}

	routeHR3 := &route{
		Source: hr3,
		ValidSectionNameRefs: map[string]struct{}{
			"listener-443-1": {},
		},
		InvalidSectionNameRefs: map[string]struct{}{},
	}

	expected := &graph{
		GatewayClass: &gatewayClass{
			Source: store.gc,
			Valid:  true,
		},
		Gateway: &gateway{
			Source: gw1,
			Listeners: map[string]*listener{
				"listener-80-1": {
					Source: gw1.Spec.Listeners[0],
					Valid:  true,
					Routes: map[types.NamespacedName]*route{
						{Namespace: "test", Name: "hr-1"}: routeHR1,
					},
					AcceptedHostnames: map[string]struct{}{
						"foo.example.com": {},
					},
				},
				"listener-443-1": {
					Source: gw1.Spec.Listeners[1],
					Valid:  true,
					Routes: map[types.NamespacedName]*route{
						{Namespace: "test", Name: "hr-3"}: routeHR3,
					},
					AcceptedHostnames: map[string]struct{}{
						"foo.example.com": {},
					},
					SecretPath: secretPath,
				},
			},
		},
		IgnoredGateways: map[types.NamespacedName]*v1alpha2.Gateway{
			{Namespace: "test", Name: "gateway-2"}: gw2,
		},
		Routes: map[types.NamespacedName]*route{
			{Namespace: "test", Name: "hr-1"}: routeHR1,
			{Namespace: "test", Name: "hr-3"}: routeHR3,
		},
	}

	// add test secret to store
	secretStore := NewSecretStore()
	secretStore.Upsert(testSecret)

	secretMemoryMgr := NewSecretDiskMemoryManager(secretsDirectory, secretStore)

	result := buildGraph(store, controllerName, gcName, secretMemoryMgr)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("buildGraph() mismatch (-want +got):\n%s", diff)
	}
}

func TestProcessGateways(t *testing.T) {
	const gcName = "test-gc"

	winner := &v1alpha2.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway-1",
		},
		Spec: v1alpha2.GatewaySpec{
			GatewayClassName: gcName,
		},
	}
	loser := &v1alpha2.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway-2",
		},
		Spec: v1alpha2.GatewaySpec{
			GatewayClassName: gcName,
		},
	}

	tests := []struct {
		gws                map[types.NamespacedName]*v1alpha2.Gateway
		expectedWinner     *v1alpha2.Gateway
		expectedIgnoredGws map[types.NamespacedName]*v1alpha2.Gateway
		msg                string
	}{
		{
			gws:                nil,
			expectedWinner:     nil,
			expectedIgnoredGws: nil,
			msg:                "no gateways",
		},
		{
			gws: map[types.NamespacedName]*v1alpha2.Gateway{
				{Namespace: "test", Name: "some-gateway"}: {
					Spec: v1alpha2.GatewaySpec{GatewayClassName: "some-class"},
				},
			},
			expectedWinner:     nil,
			expectedIgnoredGws: nil,
			msg:                "unrelated gateway",
		},
		{
			gws: map[types.NamespacedName]*v1alpha2.Gateway{
				{Namespace: "test", Name: "gateway"}: winner,
			},
			expectedWinner:     winner,
			expectedIgnoredGws: map[types.NamespacedName]*v1alpha2.Gateway{},
			msg:                "one gateway",
		},
		{
			gws: map[types.NamespacedName]*v1alpha2.Gateway{
				{Namespace: "test", Name: "gateway-1"}: winner,
				{Namespace: "test", Name: "gateway-2"}: loser,
			},
			expectedWinner: winner,
			expectedIgnoredGws: map[types.NamespacedName]*v1alpha2.Gateway{
				{Namespace: "test", Name: "gateway-2"}: loser,
			},
			msg: "multiple gateways",
		},
	}

	for _, test := range tests {
		winner, ignoredGws := processGateways(test.gws, gcName)

		if diff := cmp.Diff(winner, test.expectedWinner); diff != "" {
			t.Errorf("processGateways() '%s' mismatch for winner (-want +got):\n%s", test.msg, diff)
		}
		if diff := cmp.Diff(ignoredGws, test.expectedIgnoredGws); diff != "" {
			t.Errorf("processGateways() '%s' mismatch for ignored gateways (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestBuildGatewayClass(t *testing.T) {
	const controllerName = "my.controller"

	validGC := &v1alpha2.GatewayClass{
		Spec: v1alpha2.GatewayClassSpec{
			ControllerName: "my.controller",
		},
	}
	invalidGC := &v1alpha2.GatewayClass{
		Spec: v1alpha2.GatewayClassSpec{
			ControllerName: "wrong.controller",
		},
	}

	tests := []struct {
		gc       *v1alpha2.GatewayClass
		expected *gatewayClass
		msg      string
	}{
		{
			gc:       nil,
			expected: nil,
			msg:      "no gatewayclass",
		},
		{
			gc: validGC,
			expected: &gatewayClass{
				Source:   validGC,
				Valid:    true,
				ErrorMsg: "",
			},
			msg: "valid gatewayclass",
		},
		{
			gc: invalidGC,
			expected: &gatewayClass{
				Source:   invalidGC,
				Valid:    false,
				ErrorMsg: "Spec.ControllerName must be my.controller got wrong.controller",
			},
			msg: "invalid gatewayclass",
		},
	}

	for _, test := range tests {
		result := buildGatewayClass(test.gc, controllerName)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("buildGatewayClass() '%s' mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestBuildListeners(t *testing.T) {
	const gcName = "my-gateway-class"

	listener801 := v1alpha2.Listener{
		Name:     "listener-80-1",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     80,
		Protocol: v1alpha2.HTTPProtocolType,
	}
	listener802 := v1alpha2.Listener{
		Name:     "listener-80-2",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("bar.example.com")),
		Port:     80,
		Protocol: v1alpha2.TCPProtocolType, // invalid protocol
	}
	listener803 := v1alpha2.Listener{
		Name:     "listener-80-3",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("bar.example.com")),
		Port:     80,
		Protocol: v1alpha2.HTTPProtocolType,
	}
	listener804 := v1alpha2.Listener{
		Name:     "listener-80-4",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     80,
		Protocol: v1alpha2.HTTPProtocolType,
	}

	gatewayTLSConfig := &v1alpha2.GatewayTLSConfig{
		Mode: helpers.GetTLSModePointer(v1alpha2.TLSModeTerminate),
		CertificateRefs: []*v1alpha2.SecretObjectReference{
			{
				Kind:      (*v1alpha2.Kind)(helpers.GetStringPointer("Secret")),
				Name:      "secret",
				Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
			},
		},
	}
	// https listeners
	listener4431 := v1alpha2.Listener{
		Name:     "listener-443-1",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     443,
		TLS:      gatewayTLSConfig,
		Protocol: v1alpha2.HTTPSProtocolType,
	}
	listener4432 := v1alpha2.Listener{
		Name:     "listener-443-2",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("bar.example.com")),
		Port:     443,
		TLS:      gatewayTLSConfig,
		Protocol: v1alpha2.HTTPSProtocolType,
	}
	listener4433 := v1alpha2.Listener{
		Name:     "listener-443-3",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     443,
		TLS:      gatewayTLSConfig,
		Protocol: v1alpha2.HTTPSProtocolType,
	}
	listener4434 := v1alpha2.Listener{
		Name:     "listener-443-4",
		Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     443,
		TLS:      nil, // invalid https listener; missing tls config
		Protocol: v1alpha2.HTTPSProtocolType,
	}
	tests := []struct {
		gateway  *v1alpha2.Gateway
		expected map[string]*listener
		msg      string
	}{
		{
			gateway: &v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1alpha2.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1alpha2.Listener{
						listener801,
					},
				},
				Status: v1alpha2.GatewayStatus{},
			},
			expected: map[string]*listener{
				"listener-80-1": {
					Source:            listener801,
					Valid:             true,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
			},
			msg: "valid http listener",
		},
		{
			gateway: &v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1alpha2.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1alpha2.Listener{
						listener4431,
					},
				},
			},
			expected: map[string]*listener{
				"listener-443-1": {
					Source:            listener4431,
					Valid:             true,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
					SecretPath:        secretPath,
				},
			},
			msg: "valid https listener",
		},
		{
			gateway: &v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1alpha2.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1alpha2.Listener{
						listener802,
					},
				},
			},
			expected: map[string]*listener{
				"listener-80-2": {
					Source:            listener802,
					Valid:             false,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
			},
			msg: "invalid listener protocol",
		},
		{
			gateway: &v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1alpha2.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1alpha2.Listener{
						listener4434,
					},
				},
			},
			expected: map[string]*listener{
				"listener-443-4": {
					Source:            listener4434,
					Valid:             false,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
			},
			msg: "invalid https listener (tls config missing)",
		},
		{
			gateway: &v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1alpha2.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1alpha2.Listener{
						listener801, listener803,
						listener4431, listener4432,
					},
				},
			},
			expected: map[string]*listener{
				"listener-80-1": {
					Source:            listener801,
					Valid:             true,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
				"listener-80-3": {
					Source:            listener803,
					Valid:             true,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
				"listener-443-1": {
					Source:            listener4431,
					Valid:             true,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
					SecretPath:        secretPath,
				},
				"listener-443-2": {
					Source:            listener4432,
					Valid:             true,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
					SecretPath:        secretPath,
				},
			},
			msg: "multiple valid http/https listeners",
		},
		{
			gateway: &v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1alpha2.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1alpha2.Listener{
						listener801, listener804,
						listener4431, listener4433,
					},
				},
			},
			expected: map[string]*listener{
				"listener-80-1": {
					Source:            listener801,
					Valid:             false,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
				"listener-80-4": {
					Source:            listener804,
					Valid:             false,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
				},
				"listener-443-1": {
					Source:            listener4431,
					Valid:             false,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
					SecretPath:        secretPath,
				},
				"listener-443-3": {
					Source:            listener4433,
					Valid:             false,
					Routes:            map[types.NamespacedName]*route{},
					AcceptedHostnames: map[string]struct{}{},
					SecretPath:        secretPath,
				},
			},
			msg: "collisions",
		},
		{
			gateway:  nil,
			expected: map[string]*listener{},
			msg:      "no gateway",
		},
		{
			gateway: &v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1alpha2.GatewaySpec{
					GatewayClassName: "wrong-class",
					Listeners: []v1alpha2.Listener{
						listener801, listener804,
					},
				},
			},
			expected: map[string]*listener{},
			msg:      "wrong gatewayclass",
		},
	}

	// add secret to store
	secretStore := NewSecretStore()
	secretStore.Upsert(testSecret)

	secretMemoryMgr := NewSecretDiskMemoryManager(secretsDirectory, secretStore)

	for _, test := range tests {
		result := buildListeners(test.gateway, gcName, secretMemoryMgr)

		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("buildListeners() %q  mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestBindRouteToListeners(t *testing.T) {
	createRoute := func(hostname string, parentRefs ...v1alpha2.ParentRef) *v1alpha2.HTTPRoute {
		return &v1alpha2.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "hr-1",
			},
			Spec: v1alpha2.HTTPRouteSpec{
				CommonRouteSpec: v1alpha2.CommonRouteSpec{
					ParentRefs: parentRefs,
				},
				Hostnames: []v1alpha2.Hostname{
					v1alpha2.Hostname(hostname),
				},
			},
		}
	}

	hrNonExistingSectionName := createRoute("foo.example.com", v1alpha2.ParentRef{
		Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
		Name:        "gateway",
		SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-2")),
	})

	hrEmptySectionName := createRoute("foo.example.com", v1alpha2.ParentRef{
		Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
		Name:      "gateway",
	})

	hrIgnoredGateway := createRoute("foo.example.com", v1alpha2.ParentRef{
		Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
		Name:        "ignored-gateway",
		SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-1")),
	})

	hrFoo := createRoute("foo.example.com", v1alpha2.ParentRef{
		Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
		Name:        "gateway",
		SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-1")),
	})

	hrFooImplicitNamespace := createRoute("foo.example.com", v1alpha2.ParentRef{
		Name:        "gateway",
		SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-1")),
	})

	hrBar := createRoute("bar.example.com", v1alpha2.ParentRef{
		Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
		Name:        "gateway",
		SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-1")),
	})

	// we create a new listener each time because the function under test can modify it
	createListener := func() *listener {
		return &listener{
			Source: v1alpha2.Listener{
				Hostname: (*v1alpha2.Hostname)(helpers.GetStringPointer("foo.example.com")),
			},
			Valid:             true,
			Routes:            map[types.NamespacedName]*route{},
			AcceptedHostnames: map[string]struct{}{},
		}
	}

	createModifiedListener := func(m func(*listener)) *listener {
		l := createListener()
		m(l)
		return l
	}

	gw := &v1alpha2.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway",
		},
	}

	tests := []struct {
		httpRoute         *v1alpha2.HTTPRoute
		gw                *v1alpha2.Gateway
		ignoredGws        map[types.NamespacedName]*v1alpha2.Gateway
		listeners         map[string]*listener
		expectedIgnored   bool
		expectedRoute     *route
		expectedListeners map[string]*listener
		msg               string
	}{
		{
			httpRoute:  createRoute("foo.example.com"),
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: true,
			expectedRoute:   nil,
			expectedListeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute without parent refs",
		},
		{
			httpRoute: createRoute("foo.example.com", v1alpha2.ParentRef{
				Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
				Name:        "some-gateway", // wrong gateway
				SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-1")),
			}),
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: true,
			expectedRoute:   nil,
			expectedListeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute without good parent refs",
		},
		{
			httpRoute:  hrNonExistingSectionName,
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &route{
				Source:               hrNonExistingSectionName,
				ValidSectionNameRefs: map[string]struct{}{},
				InvalidSectionNameRefs: map[string]struct{}{
					"listener-80-2": {},
				},
			},
			expectedListeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute with non-existing section name",
		},
		{
			httpRoute:  hrEmptySectionName,
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: true,
			expectedRoute:   nil,
			expectedListeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute with empty section name",
		},
		{
			httpRoute:  hrFoo,
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &route{
				Source: hrFoo,
				ValidSectionNameRefs: map[string]struct{}{
					"listener-80-1": {},
				},
				InvalidSectionNameRefs: map[string]struct{}{},
			},
			expectedListeners: map[string]*listener{
				"listener-80-1": createModifiedListener(func(l *listener) {
					l.Routes = map[types.NamespacedName]*route{
						{Namespace: "test", Name: "hr-1"}: {
							Source: hrFoo,
							ValidSectionNameRefs: map[string]struct{}{
								"listener-80-1": {},
							},
							InvalidSectionNameRefs: map[string]struct{}{},
						},
					}
					l.AcceptedHostnames = map[string]struct{}{
						"foo.example.com": {},
					}
				}),
			},
			msg: "HTTPRoute with one accepted hostname",
		},
		{
			httpRoute:  hrFooImplicitNamespace,
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &route{
				Source: hrFooImplicitNamespace,
				ValidSectionNameRefs: map[string]struct{}{
					"listener-80-1": {},
				},
				InvalidSectionNameRefs: map[string]struct{}{},
			},
			expectedListeners: map[string]*listener{
				"listener-80-1": createModifiedListener(func(l *listener) {
					l.Routes = map[types.NamespacedName]*route{
						{Namespace: "test", Name: "hr-1"}: {
							Source: hrFooImplicitNamespace,
							ValidSectionNameRefs: map[string]struct{}{
								"listener-80-1": {},
							},
							InvalidSectionNameRefs: map[string]struct{}{},
						},
					}
					l.AcceptedHostnames = map[string]struct{}{
						"foo.example.com": {},
					}
				}),
			},
			msg: "HTTPRoute with one accepted hostname with implicit namespace in parentRef",
		},
		{
			httpRoute:  hrBar,
			gw:         gw,
			ignoredGws: nil,
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &route{
				Source:               hrBar,
				ValidSectionNameRefs: map[string]struct{}{},
				InvalidSectionNameRefs: map[string]struct{}{
					"listener-80-1": {},
				},
			},
			expectedListeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute with zero accepted hostnames",
		},
		{
			httpRoute: hrIgnoredGateway,
			gw:        gw,
			ignoredGws: map[types.NamespacedName]*v1alpha2.Gateway{
				{Namespace: "test", Name: "ignored-gateway"}: {},
			},
			listeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			expectedIgnored: false,
			expectedRoute: &route{
				Source:               hrIgnoredGateway,
				ValidSectionNameRefs: map[string]struct{}{},
				InvalidSectionNameRefs: map[string]struct{}{
					"listener-80-1": {},
				},
			},
			expectedListeners: map[string]*listener{
				"listener-80-1": createListener(),
			},
			msg: "HTTPRoute with ignored gateway reference",
		},
		{
			httpRoute:         hrFoo,
			gw:                nil,
			ignoredGws:        nil,
			listeners:         nil,
			expectedIgnored:   true,
			expectedRoute:     nil,
			expectedListeners: nil,
			msg:               "HTTPRoute when no gateway exists",
		},
	}

	for _, test := range tests {
		ignored, route := bindHTTPRouteToListeners(test.httpRoute, test.gw, test.ignoredGws, test.listeners)
		if diff := cmp.Diff(test.expectedIgnored, ignored); diff != "" {
			t.Errorf("bindHTTPRouteToListeners() %q  mismatch on ignored (-want +got):\n%s", test.msg, diff)
		}
		if diff := cmp.Diff(test.expectedRoute, route); diff != "" {
			t.Errorf("bindHTTPRouteToListeners() %q  mismatch on route (-want +got):\n%s", test.msg, diff)
		}
		if diff := cmp.Diff(test.expectedListeners, test.listeners); diff != "" {
			t.Errorf("bindHTTPRouteToListeners() %q  mismatch on listeners (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestFindAcceptedHostnames(t *testing.T) {
	var listenerHostnameFoo v1alpha2.Hostname = "foo.example.com"
	var listenerHostnameCafe v1alpha2.Hostname = "cafe.example.com"
	routeHostnames := []v1alpha2.Hostname{"foo.example.com", "bar.example.com"}

	tests := []struct {
		listenerHostname *v1alpha2.Hostname
		routeHostnames   []v1alpha2.Hostname
		expected         []string
		msg              string
	}{
		{
			listenerHostname: &listenerHostnameFoo,
			routeHostnames:   routeHostnames,
			expected:         []string{"foo.example.com"},
			msg:              "one match",
		},
		{
			listenerHostname: &listenerHostnameCafe,
			routeHostnames:   routeHostnames,
			expected:         nil,
			msg:              "no match",
		},
		{
			listenerHostname: nil,
			routeHostnames:   routeHostnames,
			expected:         []string{"foo.example.com", "bar.example.com"},
			msg:              "nil listener hostname",
		},
	}

	for _, test := range tests {
		result := findAcceptedHostnames(test.listenerHostname, test.routeHostnames)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("findAcceptedHostnames() %q  mismatch (-want +got):\n%s", test.msg, diff)
		}
	}

}

func TestGetHostname(t *testing.T) {
	var emptyHostname v1alpha2.Hostname
	var hostname v1alpha2.Hostname = "example.com"

	tests := []struct {
		h        *v1alpha2.Hostname
		expected string
		msg      string
	}{
		{
			h:        nil,
			expected: "",
			msg:      "nil hostname",
		},
		{
			h:        &emptyHostname,
			expected: "",
			msg:      "empty hostname",
		},
		{
			h:        &hostname,
			expected: string(hostname),
			msg:      "normal hostname",
		},
	}

	for _, test := range tests {
		result := getHostname(test.h)
		if result != test.expected {
			t.Errorf("getHostname() returned %q but expected %q for the case of %q", result, test.expected, test.msg)
		}
	}
}

func TestValidateGatewayClass(t *testing.T) {
	gc := &v1alpha2.GatewayClass{
		Spec: v1alpha2.GatewayClassSpec{
			ControllerName: "test.controller",
		},
	}

	err := validateGatewayClass(gc, "test.controller")
	if err != nil {
		t.Errorf("validateGatewayClass() returned unexpected error %v", err)
	}

	err = validateGatewayClass(gc, "unmatched.controller")
	if err == nil {
		t.Errorf("validateGatewayClass() didn't return an error")
	}
}
