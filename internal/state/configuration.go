package state

import (
	"fmt"
	"sort"

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// Configuration is an internal representation of Gateway configuration.
// We can think of Configuration as an intermediate state between the Gateway API resources and the data plane (NGINX)
// configuration.
type Configuration struct {
	// HTTPServers holds all HTTPServers.
	// FIXME(pleshakov) We assume that all servers are HTTP and listen on port 80.
	HTTPServers []VirtualServer
	// SSLServers holds all SSLServers.
	// FIXME(kate-osborn) We assume that all SSL servers listen on port 443.
	SSLServers []VirtualServer
}

// VirtualServer is a virtual server.
type VirtualServer struct {
	// Hostname is the hostname of the server.
	Hostname string
	// PathRules is a collection of routing rules.
	PathRules []PathRule
	// SSL holds the SSL configuration options fo the server.
	SSL *SSL
}

type SSL struct {
	// CertificatePath is the path to the certificate file.
	CertificatePath string
}

// PathRule represents routing rules that share a common path.
type PathRule struct {
	// Path is a path. For example, '/hello'.
	Path string
	// MatchRules holds routing rules.
	MatchRules []MatchRule
}

// MatchRule represents a routing rule. It corresponds directly to a Match in the HTTPRoute resource.
// An HTTPRoute is guaranteed to have at least one rule with one match.
// If no rule or match is specified by the user, the default rule {{path:{ type: "PathPrefix", value: "/"}}} is set by the schema.
type MatchRule struct {
	// MatchIdx is the index of the rule in the Rule.Matches.
	MatchIdx int
	// RuleIdx is the index of the corresponding rule in the HTTPRoute.
	RuleIdx int
	// Source is the corresponding HTTPRoute resource.
	Source *v1alpha2.HTTPRoute
}

// GetMatch returns the HTTPRouteMatch of the Route .
func (r *MatchRule) GetMatch() v1alpha2.HTTPRouteMatch {
	return r.Source.Spec.Rules[r.RuleIdx].Matches[r.MatchIdx]
}

// buildConfiguration builds the Configuration from the graph.
// FIXME(pleshakov) For now we only handle paths with prefix matches. Handle exact and regex matches
func buildConfiguration(graph *graph) Configuration {
	if graph.GatewayClass == nil || !graph.GatewayClass.Valid {
		return Configuration{}
	}

	if graph.Gateway == nil {
		return Configuration{}
	}

	configBuilder := newConfigBuilder()

	for _, l := range graph.Gateway.Listeners {
		// only upsert listeners that are valid
		if l.Valid {
			configBuilder.upsertListener(l)
		}
	}

	return configBuilder.build()
}

type configBuilder struct {
	http *virtualServerBuilder
	ssl  *virtualServerBuilder
}

func newConfigBuilder() *configBuilder {
	return &configBuilder{
		http: newVirtualServerBuilder(),
		ssl:  newVirtualServerBuilder(),
	}
}

func (b *configBuilder) upsertListener(l *listener) {
	switch l.Source.Protocol {
	case v1alpha2.HTTPProtocolType:
		b.http.upsertListener(l)
	case v1alpha2.HTTPSProtocolType:
		b.ssl.upsertListener(l)
	default:
		panic(fmt.Sprintf("listener protocol %s not supported", l.Source.Protocol))
	}
}

func (b *configBuilder) build() Configuration {
	return Configuration{
		HTTPServers: b.http.build(),
		SSLServers:  b.ssl.build(),
	}
}

type virtualServerBuilder struct {
	rulesPerHost     map[string]map[string]PathRule
	listenersForHost map[string]*listener
}

func newVirtualServerBuilder() *virtualServerBuilder {
	return &virtualServerBuilder{
		rulesPerHost:     make(map[string]map[string]PathRule),
		listenersForHost: make(map[string]*listener),
	}
}

func (b *virtualServerBuilder) upsertListener(l *listener) {

	for _, r := range l.Routes {
		var hostnames []string

		for _, h := range r.Source.Spec.Hostnames {
			if _, exist := l.AcceptedHostnames[string(h)]; exist {
				hostnames = append(hostnames, string(h))
			}
		}

		for _, h := range hostnames {
			b.listenersForHost[h] = l
			if _, exist := b.rulesPerHost[h]; !exist {
				b.rulesPerHost[h] = make(map[string]PathRule)
			}
		}

		for i, rule := range r.Source.Spec.Rules {
			for _, h := range hostnames {
				for j, m := range rule.Matches {
					path := getPath(m.Path)

					rule, exist := b.rulesPerHost[h][path]
					if !exist {
						rule.Path = path
					}

					rule.MatchRules = append(rule.MatchRules, MatchRule{
						MatchIdx: j,
						RuleIdx:  i,
						Source:   r.Source,
					})

					b.rulesPerHost[h][path] = rule
				}
			}
		}
	}
}

func (b *virtualServerBuilder) build() []VirtualServer {

	servers := make([]VirtualServer, 0, len(b.rulesPerHost))

	for h, rules := range b.rulesPerHost {
		s := VirtualServer{
			Hostname:  h,
			PathRules: make([]PathRule, 0, len(rules)),
		}

		l, ok := b.listenersForHost[h]
		if !ok {
			panic(fmt.Sprintf("no listener found for hostname: %s", h))
		}

		if l.SecretPath != "" {
			s.SSL = &SSL{CertificatePath: l.SecretPath}
		}

		for _, r := range rules {
			sortMatchRules(r.MatchRules)

			s.PathRules = append(s.PathRules, r)
		}

		// sort rules for predictable order
		sort.Slice(s.PathRules, func(i, j int) bool {
			return s.PathRules[i].Path < s.PathRules[j].Path
		})

		servers = append(servers, s)
	}

	// sort servers for predictable order
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Hostname < servers[j].Hostname
	})

	return servers
}

func getPath(path *v1alpha2.HTTPPathMatch) string {
	if path == nil || path.Value == nil || *path.Value == "" {
		return "/"
	}
	return *path.Value
}
