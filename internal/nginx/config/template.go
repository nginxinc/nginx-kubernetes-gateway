package config

import (
	"bytes"
	"fmt"
	"text/template"
)

var httpServersTemplate = `{{ range $s := .Servers }}
	{{ if $s.IsDefaultSSL }}
server {
	listen 443 ssl default_server;

	ssl_reject_handshake on;
}
	{{ else if $s.IsDefaultHTTP }}
server {
	listen 80 default_server;
	
	default_type text/html;
	return 404;
}
	{{ else }}
server {
	{{ if $s.SSL }}
	listen 443 ssl;
	ssl_certificate {{ $s.SSL.Certificate }};
	ssl_certificate_key {{ $s.SSL.CertificateKey }};
	{{ end }}

	server_name {{ $s.ServerName }};

	{{ range $l := $s.Locations }}
	location {{ $l.Path }} {
		{{ if $l.Internal }}
		internal;
		{{ end }}
		
		proxy_set_header Host $host;

		{{ if $l.HTTPMatchVar }}
		set $http_matches {{ $l.HTTPMatchVar | printf "%q" }};
		js_content httpmatches.redirect;
		{{ end }}

		{{ if $l.ProxyPass }}
		proxy_pass {{ $l.ProxyPass }}$request_uri;
		{{ end }}
	}
	{{ end }}
}
	{{ end }}
{{ end }}
`

// templateExecutor generates NGINX configuration using a template.
// Template parsing or executing errors can only occur if there is a bug in the template, so they are handled with panics.
// For now, we only generate configuration with NGINX http servers, but in the future we will also need to generate
// the main NGINX configuration file, upstreams, stream servers.
type templateExecutor struct {
	httpServersTemplate *template.Template
}

func newTemplateExecutor() *templateExecutor {
	t, err := template.New("server").Parse(httpServersTemplate)
	if err != nil {
		panic(fmt.Errorf("failed to parse http servers template: %w", err))
	}

	return &templateExecutor{httpServersTemplate: t}
}

func (e *templateExecutor) ExecuteForHTTPServers(servers httpServers) []byte {
	var buf bytes.Buffer

	err := e.httpServersTemplate.Execute(&buf, servers)
	if err != nil {
		panic(fmt.Errorf("failed to execute http servers template: %w", err))
	}

	return buf.Bytes()
}
