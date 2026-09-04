package domain

import (
	"strings"
	"time"
)

// Domain maps a public fully qualified domain name (FQDN) to a service.
type Domain struct {
	ID          string    `json:"id"`
	ServiceID   string    `json:"serviceId"`
	ProjectName string    `json:"projectName,omitempty"`
	ServiceName string    `json:"serviceName,omitempty"`
	DomainName  string    `json:"domainName"`
	Port        int       `json:"port"`
	Path        string    `json:"path,omitempty"`
	HTTPS       bool      `json:"https"`
	CertMode    string    `json:"certMode,omitempty"` // "letsencrypt" or "custom" or "none"
	Middlewares []string  `json:"middlewares,omitempty"`
	Status      string    `json:"status"` // "active", "pending", "error"
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Validate ensures domain fields are valid.
func (d *Domain) Validate() error {
	if strings.TrimSpace(d.ID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(d.ServiceID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(d.DomainName) == "" {
		return ErrValidation
	}
	if d.Port <= 0 || d.Port > 65535 {
		return ErrValidation
	}
	return nil
}

// ToRouteConfig converts a domain entity to reverse proxy RouteConfig.
func (d *Domain) ToRouteConfig() RouteConfig {
	certResolver := ""
	if d.HTTPS && (d.CertMode == "letsencrypt" || d.CertMode == "") {
		certResolver = "letsencrypt"
	}
	return RouteConfig{
		ServiceID:    d.ServiceID,
		Domain:       d.DomainName,
		TargetPort:   d.Port,
		PathPrefix:   d.Path,
		EnableHTTPS:  d.HTTPS,
		CertResolver: certResolver,
		Middlewares:  d.Middlewares,
	}
}
