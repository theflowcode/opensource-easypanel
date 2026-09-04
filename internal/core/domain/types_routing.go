package domain

// RouteConfig specifies reverse proxy routing for Traefik.
type RouteConfig struct {
	ServiceID      string   `json:"serviceId"`
	Domain         string   `json:"domain"`
	TargetPort     int      `json:"targetPort"`
	PathPrefix     string   `json:"pathPrefix,omitempty"`
	EnableHTTPS    bool     `json:"enableHttps"`
	CertResolver   string   `json:"certResolver,omitempty"` // "letsencrypt"
	Middlewares    []string `json:"middlewares,omitempty"`
	PassHostHeader bool     `json:"passHostHeader"`
}

// RedirectRule defines an HTTP redirect rule for a service (Domains -> Redirects tab).
type RedirectRule struct {
	ID          string `json:"id"`
	ServiceID   string `json:"serviceId"`
	Source      string `json:"source,omitempty"`      // Source domain or path e.g. "www.example.com"
	Target      string `json:"target,omitempty"`      // Target URL or path e.g. "https://example.com"
	Regex       string `json:"regex,omitempty"`       // Regex path pattern e.g. "^/old/(.*)"
	Replacement string `json:"replacement,omitempty"` // Replacement target e.g. "/new/$1"
	Permanent   bool   `json:"permanent"`            // true = 301 Moved Permanently, false = 302 Found
	Enabled     bool   `json:"enabled"`
}
