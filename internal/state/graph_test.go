package state

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func TestBuildGraph(t *testing.T) {
	const (
		gcName         = "my-class"
		controllerName = "my.controller"
	)
	createRoute := func(name string, gatewayName string) *v1alpha2.HTTPRoute {
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
							SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-1")),
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
	hr1 := createRoute("hr-1", "gateway-1")
	hr2 := createRoute("hr-2", "wrong-gateway")

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
		},
	}

	routeHR1 := &route{
		Source: hr1,
		ValidSectionNameRefs: map[string]struct{}{
			"listener-80-1": {},
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
			},
		},
		IgnoredGateways: map[types.NamespacedName]*v1alpha2.Gateway{
			{Namespace: "test", Name: "gateway-2"}: gw2,
		},
		Routes: map[types.NamespacedName]*route{
			{Namespace: "test", Name: "hr-1"}: routeHR1,
		},
	}

	result := buildGraph(store, controllerName, gcName)
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

	tests := []struct {
		gateway  *v1alpha2.Gateway
		expected map[string]*listener
		msg      string
	}{
		{
			gateway: &v1alpha2.Gateway{
				Spec: v1alpha2.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1alpha2.Listener{
						listener801,
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
			},
			msg: "valid listener",
		},
		{
			gateway: &v1alpha2.Gateway{
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
			msg: "invalid listener",
		},
		{
			gateway: &v1alpha2.Gateway{
				Spec: v1alpha2.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1alpha2.Listener{
						listener801, listener803,
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
			},
			msg: "two valid Listeners",
		},
		{
			gateway: &v1alpha2.Gateway{
				Spec: v1alpha2.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1alpha2.Listener{
						listener801, listener804,
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
			},
			msg: "collision",
		},
		{
			gateway:  nil,
			expected: map[string]*listener{},
			msg:      "no gateway",
		},
		{
			gateway: &v1alpha2.Gateway{
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

	for _, test := range tests {
		result := buildListeners(test.gateway, gcName)
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

func TestValidateListener(t *testing.T) {
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
		{
			l: v1alpha2.Listener{
				Port:     80,
				Protocol: v1alpha2.TCPProtocolType,
			},
			expected: false,
			msg:      "invalid protocol",
		},
	}

	for _, test := range tests {
		result := validateListener(test.l)
		if result != test.expected {
			t.Errorf("validateListener() returned %v but expected %v for the case of %q", result, test.expected, test.msg)
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
