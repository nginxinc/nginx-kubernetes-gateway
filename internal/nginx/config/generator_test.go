package config

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/statefakes"
)

func TestGenerateForHost(t *testing.T) {
	generator := NewGeneratorImpl(&statefakes.FakeServiceStore{})

	testcases := []struct {
		conf        state.Configuration
		httpDefault bool
		sslDefault  bool
		msg         string
	}{
		{
			conf:        state.Configuration{},
			httpDefault: false,
			sslDefault:  false,
			msg:         "no servers",
		},
		{
			conf: state.Configuration{
				HTTPServers: []state.VirtualServer{
					{
						Hostname: "example.com",
					},
				},
			},
			httpDefault: true,
			sslDefault:  false,
			msg:         "only HTTP servers",
		},
		{
			conf: state.Configuration{
				SSLServers: []state.VirtualServer{
					{
						Hostname: "example.com",
					},
				},
			},
			httpDefault: false,
			sslDefault:  true,
			msg:         "only HTTPS servers",
		},
		{
			conf: state.Configuration{
				HTTPServers: []state.VirtualServer{
					{
						Hostname: "example.com",
					},
				},
				SSLServers: []state.VirtualServer{
					{
						Hostname: "example.com",
					},
				},
			},
			httpDefault: true,
			sslDefault:  true,
			msg:         "both HTTP and HTTPS servers",
		},
	}

	for _, tc := range testcases {
		cfg, warnings := generator.Generate(tc.conf)

		defaultSSLExists := strings.Contains(string(cfg), "listen 443 ssl default_server")
		defaultHTTPExists := strings.Contains(string(cfg), "listen 80 default_server")

		if tc.sslDefault && !defaultSSLExists {
			t.Errorf("Generate() did not generate a config with a default TLS termination server for test: %q", tc.msg)
		}

		if !tc.sslDefault && defaultSSLExists {
			t.Errorf("Generate() generated a config with a default TLS termination server for test: %q", tc.msg)
		}

		if tc.httpDefault && !defaultHTTPExists {
			t.Errorf("Generate() did not generate a config with a default http server for test: %q", tc.msg)
		}

		if !tc.httpDefault && defaultHTTPExists {
			t.Errorf("Generate() generated a config with a default http server for test: %q", tc.msg)
		}

		if len(cfg) == 0 {
			t.Errorf("Generate() generated empty config for test: %q", tc.msg)
		}
		if len(warnings) > 0 {
			t.Errorf("Generate() returned unexpected warnings: %v for test: %q", warnings, tc.msg)
		}
	}
}

func TestGenerate(t *testing.T) {
	hr := &v1alpha2.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "route1",
		},
		Spec: v1alpha2.HTTPRouteSpec{
			Hostnames: []v1alpha2.Hostname{
				"cafe.example.com",
			},
			Rules: []v1alpha2.HTTPRouteRule{
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/"),
							},
							Method: helpers.GetHTTPMethodPointer(v1alpha2.HTTPMethodPost),
						},
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/"),
							},
							Method: helpers.GetHTTPMethodPointer(v1alpha2.HTTPMethodPatch),
						},
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/"), // should generate an "any" httpmatch since other matches exists for /
							},
						},
					},
					BackendRefs: []v1alpha2.HTTPBackendRef{
						{
							BackendRef: v1alpha2.BackendRef{
								BackendObjectReference: v1alpha2.BackendObjectReference{
									Name:      "service1",
									Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
									Port:      (*v1alpha2.PortNumber)(helpers.GetInt32Pointer(80)),
								},
							},
						},
					},
				},
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/test"),
							},
							Method: helpers.GetHTTPMethodPointer(v1alpha2.HTTPMethodGet),
							Headers: []v1alpha2.HTTPHeaderMatch{
								{
									Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
									Name:  "Version",
									Value: "V1",
								},
								{
									Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
									Name:  "test",
									Value: "foo",
								},
								{
									Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
									Name:  "my-header",
									Value: "my-value",
								},
							},
							QueryParams: []v1alpha2.HTTPQueryParamMatch{
								{
									Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchExact),
									Name:  "GrEat", // query names and values should not be normalized to lowercase
									Value: "EXAMPLE",
								},
								{
									Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchExact),
									Name:  "test",
									Value: "foo=bar",
								},
							},
						},
					},
					BackendRefs: nil, // no backend refs will cause warnings
				},
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-only"),
							},
						},
					},
					BackendRefs: []v1alpha2.HTTPBackendRef{
						{
							BackendRef: v1alpha2.BackendRef{
								BackendObjectReference: v1alpha2.BackendObjectReference{
									Name:      "service2",
									Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
									Port:      (*v1alpha2.PortNumber)(helpers.GetInt32Pointer(80)),
								},
							},
						},
					},
				},
			},
		},
	}

	certPath := "/etc/nginx/secrets/cert"

	httpHost := state.VirtualServer{
		Hostname: "example.com",
		PathRules: []state.PathRule{
			{
				Path: "/",
				MatchRules: []state.MatchRule{
					{
						MatchIdx: 0,
						RuleIdx:  0,
						Source:   hr,
					},
					{
						MatchIdx: 1,
						RuleIdx:  0,
						Source:   hr,
					},
					{
						MatchIdx: 2,
						RuleIdx:  0,
						Source:   hr,
					},
				},
			},
			{
				Path: "/test",
				MatchRules: []state.MatchRule{
					{
						MatchIdx: 0,
						RuleIdx:  1,
						Source:   hr,
					},
				},
			},
			{
				Path: "/path-only",
				MatchRules: []state.MatchRule{
					{
						MatchIdx: 0,
						RuleIdx:  2,
						Source:   hr,
					},
				},
			},
		},
	}

	httpsHost := httpHost
	httpsHost.SSL = &state.SSL{CertificatePath: certPath}

	fakeServiceStore := &statefakes.FakeServiceStore{}
	fakeServiceStore.ResolveReturns("10.0.0.1", nil)

	expectedMatchString := func(m []httpMatch) string {
		b, err := json.Marshal(m)
		if err != nil {
			t.Errorf("error marshaling test match: %v", err)
		}
		return string(b)
	}

	slashMatches := []httpMatch{
		{Method: v1alpha2.HTTPMethodPost, RedirectPath: "/_route0"},
		{Method: v1alpha2.HTTPMethodPatch, RedirectPath: "/_route1"},
		{Any: true, RedirectPath: "/_route2"},
	}
	testMatches := []httpMatch{
		{
			Method:       v1alpha2.HTTPMethodGet,
			Headers:      []string{"Version:V1", "test:foo", "my-header:my-value"},
			QueryParams:  []string{"GrEat=EXAMPLE", "test=foo=bar"},
			RedirectPath: "/test_route0",
		},
	}

	const backendAddr = "http://10.0.0.1:80"

	expectedHTTPServer := server{
		ServerName: "example.com",
		Locations: []location{
			{
				Path:      "/_route0",
				Internal:  true,
				ProxyPass: backendAddr,
			},
			{
				Path:      "/_route1",
				Internal:  true,
				ProxyPass: backendAddr,
			},
			{
				Path:      "/_route2",
				Internal:  true,
				ProxyPass: backendAddr,
			},
			{
				Path:         "/",
				HTTPMatchVar: expectedMatchString(slashMatches),
			},
			{
				Path:      "/test_route0",
				Internal:  true,
				ProxyPass: "http://" + nginx502Server,
			},
			{
				Path:         "/test",
				HTTPMatchVar: expectedMatchString(testMatches),
			},
			{
				Path:      "/path-only",
				ProxyPass: backendAddr,
			},
		},
	}

	expectedHTTPSServer := expectedHTTPServer
	expectedHTTPSServer.SSL = &ssl{Certificate: certPath, CertificateKey: certPath}

	expectedWarnings := Warnings{
		hr: []string{"empty backend refs"},
	}

	testcases := []struct {
		host        state.VirtualServer
		expWarnings Warnings
		expResult   server
		msg         string
	}{
		{
			host:        httpHost,
			expWarnings: expectedWarnings,
			expResult:   expectedHTTPServer,
			msg:         "http server",
		},
		{
			host:        httpsHost,
			expWarnings: expectedWarnings,
			expResult:   expectedHTTPSServer,
			msg:         "https server",
		},
	}

	for _, tc := range testcases {
		result, warnings := generate(tc.host, fakeServiceStore)

		if diff := cmp.Diff(tc.expResult, result); diff != "" {
			t.Errorf("generate() mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(tc.expWarnings, warnings); diff != "" {
			t.Errorf("generate() mismatch on warnings (-want +got):\n%s", diff)
		}
	}
}

func TestGenerateProxyPass(t *testing.T) {
	expected := "http://10.0.0.1:80"

	result := generateProxyPass("10.0.0.1:80")
	if result != expected {
		t.Errorf("generateProxyPass() returned %s but expected %s", result, expected)
	}

	expected = "http://" + nginx502Server

	result = generateProxyPass("")
	if result != expected {
		t.Errorf("generateProxyPass() returned %s but expected %s", result, expected)
	}
}

func TestGetBackendAddress(t *testing.T) {
	getNormalRefs := func() []v1alpha2.HTTPBackendRef {
		return []v1alpha2.HTTPBackendRef{
			{
				BackendRef: v1alpha2.BackendRef{
					BackendObjectReference: v1alpha2.BackendObjectReference{
						Group:     (*v1alpha2.Group)(helpers.GetStringPointer("networking.k8s.io")),
						Kind:      (*v1alpha2.Kind)(helpers.GetStringPointer("Service")),
						Name:      "service1",
						Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
						Port:      (*v1alpha2.PortNumber)(helpers.GetInt32Pointer(80)),
					},
				},
			},
		}
	}

	getModifiedRefs := func(mod func([]v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
		return mod(getNormalRefs())
	}

	tests := []struct {
		refs                      []v1alpha2.HTTPBackendRef
		parentNS                  string
		storeAddress              string
		storeErr                  error
		expectedResolverCallCount int
		expectedNsName            types.NamespacedName
		expectedAddress           string
		expectErr                 bool
		msg                       string
	}{
		{
			refs:                      getNormalRefs(),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "10.0.0.1:80",
			expectErr:                 false,
			msg:                       "normal case",
		},
		{
			refs: getModifiedRefs(
				func(refs []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
					refs[0].BackendRef.Namespace = nil
					return refs
				},
			),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "10.0.0.1:80",
			expectErr:                 false,
			msg:                       "normal case with implicit namespace",
		},
		{
			refs: getModifiedRefs(
				func(refs []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
					refs[0].BackendRef.Group = nil
					refs[0].BackendRef.Kind = nil
					return refs
				},
			),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "10.0.0.1:80",
			expectErr:                 false,
			msg:                       "normal case with implicit service",
		},
		{
			refs: getModifiedRefs(
				func(refs []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
					secondRef := refs[0].DeepCopy()
					secondRef.Name = "service2"
					return append(refs, *secondRef)
				},
			),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "10.0.0.1:80",
			expectErr:                 false,
			msg:                       "first backend ref is used",
		},
		{
			refs: getModifiedRefs(
				func(refs []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
					refs[0].BackendRef.Kind = (*v1alpha2.Kind)(helpers.GetStringPointer("NotService"))
					return refs
				},
			),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 0,
			expectedNsName:            types.NamespacedName{},
			expectedAddress:           "",
			expectErr:                 true,
			msg:                       "not a service Kind",
		},
		{
			refs:                      nil,
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 0,
			expectedNsName:            types.NamespacedName{},
			expectedAddress:           "",
			expectErr:                 true,
			msg:                       "no refs",
		},
		{
			refs: getModifiedRefs(
				func(refs []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
					refs[0].BackendRef.Port = nil
					return refs
				},
			),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "",
			expectErr:                 true,
			msg:                       "no port",
		},
		{
			refs:                      getNormalRefs(),
			parentNS:                  "test",
			storeAddress:              "",
			storeErr:                  errors.New(""),
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "",
			expectErr:                 true,
			msg:                       "service doesn't exist",
		},
	}

	for _, test := range tests {
		fakeServiceStore := &statefakes.FakeServiceStore{}
		fakeServiceStore.ResolveReturns(test.storeAddress, test.storeErr)

		result, err := getBackendAddress(test.refs, test.parentNS, fakeServiceStore)
		if result != test.expectedAddress {
			t.Errorf(
				"getBackendAddress() returned %s but expected %s for case %q",
				result,
				test.expectedAddress,
				test.msg,
			)
		}

		if test.expectErr {
			if err == nil {
				t.Errorf("getBackendAddress() didn't return any error for case %q", test.msg)
			}
		} else {
			if err != nil {
				t.Errorf("getBackendAddress() returned unexpected error %v for case %q", err, test.msg)
			}
		}

		callCount := fakeServiceStore.ResolveCallCount()
		if callCount != test.expectedResolverCallCount {
			t.Errorf(
				"getBackendAddress() called fakeServiceStore.Resolve %d times but expected %d for case %q",
				callCount,
				test.expectedResolverCallCount,
				test.msg,
			)
		}

		if test.expectedResolverCallCount == 0 {
			continue
		}

		nsname := fakeServiceStore.ResolveArgsForCall(0)
		if nsname != test.expectedNsName {
			t.Errorf(
				"getBackendAddress() called fakeServiceStore.Resolve with %v but expected %v for case %q",
				nsname,
				test.expectedNsName,
				test.msg,
			)
		}
	}
}

func TestGenerateMatchLocation(t *testing.T) {
	expected := location{
		Path:      "/path",
		Internal:  true,
		ProxyPass: "http://10.0.0.1:80",
	}

	result := generateMatchLocation("/path", "10.0.0.1:80")
	if result != expected {
		t.Errorf("generateMatchLocation() returned %v but expected %v", result, expected)
	}
}

func TestCreatePathForMatch(t *testing.T) {
	expected := "/path_route1"

	result := createPathForMatch("/path", 1)
	if result != expected {
		t.Errorf("createPathForMatch() returned %q but expected %q", result, expected)
	}
}

func TestCreateArgKeyValString(t *testing.T) {
	expected := "key=value"

	result := createQueryParamKeyValString(
		v1alpha2.HTTPQueryParamMatch{
			Name:  "key",
			Value: "value",
		},
	)
	if result != expected {
		t.Errorf("createQueryParamKeyValString() returned %q but expected %q", result, expected)
	}

	expected = "KeY=vaLUe=="

	result = createQueryParamKeyValString(
		v1alpha2.HTTPQueryParamMatch{
			Name:  "KeY",
			Value: "vaLUe==",
		},
	)
	if result != expected {
		t.Errorf("createQueryParamKeyValString() returned %q but expected %q", result, expected)
	}
}

func TestCreateHeaderKeyValString(t *testing.T) {
	expected := "kEy:vALUe"

	result := createHeaderKeyValString(
		v1alpha2.HTTPHeaderMatch{
			Name:  "kEy",
			Value: "vALUe",
		},
	)

	if result != expected {
		t.Errorf("createHeaderKeyValString() returned %q but expected %q", result, expected)
	}
}

func TestMatchLocationNeeded(t *testing.T) {
	tests := []struct {
		match    v1alpha2.HTTPRouteMatch
		expected bool
		msg      string
	}{
		{
			match: v1alpha2.HTTPRouteMatch{
				Path: &v1alpha2.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
			},
			expected: true,
			msg:      "path only match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Path: &v1alpha2.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
				Method: helpers.GetHTTPMethodPointer(v1alpha2.HTTPMethodGet),
			},
			expected: false,
			msg:      "method defined in match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Path: &v1alpha2.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
				Headers: []v1alpha2.HTTPHeaderMatch{
					{
						Name:  "header",
						Value: "val",
					},
				},
			},
			expected: false,
			msg:      "headers defined in match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{

				Path: &v1alpha2.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
				QueryParams: []v1alpha2.HTTPQueryParamMatch{
					{
						Name:  "arg",
						Value: "val",
					},
				},
			},
			expected: false,
			msg:      "query params defined in match",
		},
	}

	for _, tc := range tests {
		result := isPathOnlyMatch(tc.match)

		if result != tc.expected {
			t.Errorf("isPathOnlyMatch() returned %t but expected %t for test case %q", result, tc.expected, tc.msg)
		}
	}
}

func TestCreateHTTPMatch(t *testing.T) {
	testPath := "/internal_loc"

	testPathMatch := v1alpha2.HTTPPathMatch{Value: helpers.GetStringPointer("/")}
	testMethodMatch := helpers.GetHTTPMethodPointer(v1alpha2.HTTPMethodPut)
	testHeaderMatches := []v1alpha2.HTTPHeaderMatch{
		{
			Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
			Name:  "header-1",
			Value: "val-1",
		},
		{
			Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
			Name:  "header-2",
			Value: "val-2",
		},
		{
			// regex type is not supported. This should not be added to the httpMatch headers.
			Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchRegularExpression),
			Name:  "ignore-this-header",
			Value: "val",
		},
		{
			Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
			Name:  "header-3",
			Value: "val-3",
		},
	}

	testDuplicateHeaders := make([]v1alpha2.HTTPHeaderMatch, 0, 5)
	duplicateHeaderMatch := v1alpha2.HTTPHeaderMatch{
		Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
		Name:  "HEADER-2", // header names are case-insensitive
		Value: "val-2",
	}
	testDuplicateHeaders = append(testDuplicateHeaders, testHeaderMatches...)
	testDuplicateHeaders = append(testDuplicateHeaders, duplicateHeaderMatch)

	testQueryParamMatches := []v1alpha2.HTTPQueryParamMatch{
		{
			Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchExact),
			Name:  "arg1",
			Value: "val1",
		},
		{
			Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchExact),
			Name:  "arg2",
			Value: "val2=another-val",
		},
		{
			// regex type is not supported. This should not be added to the httpMatch args
			Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchRegularExpression),
			Name:  "ignore-this-arg",
			Value: "val",
		},
		{
			Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchExact),
			Name:  "arg3",
			Value: "==val3",
		},
	}

	expectedHeaders := []string{"header-1:val-1", "header-2:val-2", "header-3:val-3"}
	expectedArgs := []string{"arg1=val1", "arg2=val2=another-val", "arg3===val3"}

	tests := []struct {
		match    v1alpha2.HTTPRouteMatch
		expected httpMatch
		msg      string
	}{
		{
			match: v1alpha2.HTTPRouteMatch{
				Path: &testPathMatch,
			},
			expected: httpMatch{
				Any:          true,
				RedirectPath: testPath,
			},
			msg: "path only match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Path:   &testPathMatch, // A path match with a method should not set the Any field to true
				Method: testMethodMatch,
			},
			expected: httpMatch{
				Method:       "PUT",
				RedirectPath: testPath,
			},
			msg: "method only match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Headers: testHeaderMatches,
			},
			expected: httpMatch{
				RedirectPath: testPath,
				Headers:      expectedHeaders,
			},
			msg: "headers only match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				QueryParams: testQueryParamMatches,
			},
			expected: httpMatch{
				QueryParams:  expectedArgs,
				RedirectPath: testPath,
			},
			msg: "query params only match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Method:      testMethodMatch,
				QueryParams: testQueryParamMatches,
			},
			expected: httpMatch{
				Method:       "PUT",
				QueryParams:  expectedArgs,
				RedirectPath: testPath,
			},
			msg: "method and query params match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Method:  testMethodMatch,
				Headers: testHeaderMatches,
			},
			expected: httpMatch{
				Method:       "PUT",
				Headers:      expectedHeaders,
				RedirectPath: testPath,
			},
			msg: "method and headers match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				QueryParams: testQueryParamMatches,
				Headers:     testHeaderMatches,
			},
			expected: httpMatch{
				QueryParams:  expectedArgs,
				Headers:      expectedHeaders,
				RedirectPath: testPath,
			},
			msg: "query params and headers match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Headers:     testHeaderMatches,
				QueryParams: testQueryParamMatches,
				Method:      testMethodMatch,
			},
			expected: httpMatch{
				Method:       "PUT",
				Headers:      expectedHeaders,
				QueryParams:  expectedArgs,
				RedirectPath: testPath,
			},
			msg: "method, headers, and query params match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Headers: testDuplicateHeaders,
			},
			expected: httpMatch{
				Headers:      expectedHeaders,
				RedirectPath: testPath,
			},
			msg: "duplicate header names",
		},
	}
	for _, tc := range tests {
		result := createHTTPMatch(tc.match, testPath)
		if diff := helpers.Diff(result, tc.expected); diff != "" {
			t.Errorf("createHTTPMatch() returned incorrect httpMatch for test case: %q, diff: %+v", tc.msg, diff)
		}
	}
}
