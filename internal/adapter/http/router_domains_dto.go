package http

import (
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

type listDomainsInput struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
}

type createDomainInput struct {
	ProjectName         string   `json:"projectName"`
	ServiceName         string   `json:"serviceName"`
	Host                string   `json:"host"`
	DomainName          string   `json:"domainName"`
	Path                string   `json:"path"`
	Port                int      `json:"port"`
	HTTPS               bool     `json:"https"`
	CertificateResolver string   `json:"certificateResolver"`
	Middlewares         []string `json:"middlewares"`
}

type updateDomainInput struct {
	ID                  string   `json:"id"`
	Host                string   `json:"host"`
	DomainName          string   `json:"domainName"`
	Path                string   `json:"path"`
	Port                int      `json:"port"`
	HTTPS               bool     `json:"https"`
	CertificateResolver string   `json:"certificateResolver"`
	Middlewares         []string `json:"middlewares"`
}

type deleteDomainInput struct {
	ID          string `json:"id"`
	Host        string `json:"host"`
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
}

type setPrimaryDomainInput struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
	Domain      string `json:"domain"`
	ID          string `json:"id"`
}

type domainDTO struct {
	ID                  string         `json:"id"`
	Host                string         `json:"host"`
	Path                string         `json:"path"`
	Port                int            `json:"port"`
	HTTPS               bool           `json:"https"`
	Wildcard            bool           `json:"wildcard"`
	CertificateResolver string         `json:"certificateResolver"`
	DestinationType     string         `json:"destinationType"`
	ProjectName         string         `json:"projectName"`
	ServiceName         string         `json:"serviceName"`
	ServiceDestination  map[string]any `json:"serviceDestination"`
	Middlewares         []string       `json:"middlewares"`
	CreatedAt           time.Time      `json:"createdAt"`
}

func toDomainDTO(d *domain.Domain) domainDTO {
	p := d.Path
	if p == "" {
		p = "/"
	}
	portNum := d.Port
	if portNum <= 0 {
		portNum = 80
	}
	cert := d.CertMode
	if cert == "" && d.HTTPS {
		cert = "letsencrypt"
	}
	mws := d.Middlewares
	if mws == nil {
		mws = []string{}
	}

	return domainDTO{
		ID:                  d.ID,
		Host:                d.DomainName,
		Path:                p,
		Port:                portNum,
		HTTPS:               d.HTTPS,
		Wildcard:            strings.HasPrefix(d.DomainName, "*."),
		CertificateResolver: cert,
		DestinationType:     "service",
		ProjectName:         d.ProjectName,
		ServiceName:         d.ServiceName,
		ServiceDestination: map[string]any{
			"protocol":    "http",
			"projectName": d.ProjectName,
			"serviceName": d.ServiceName,
			"port":        portNum,
			"path":        p,
		},
		Middlewares: mws,
		CreatedAt:   d.CreatedAt,
	}
}
