package http

import (
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

// registerDomainsRoutes binds custom domain procedures to the oRPC dispatcher.
func registerDomainsRoutes(d *orpc.Dispatcher, db port.DatabasePort, proxy port.ProxyDriverPort) {
	d.Register("domains/listDomains", func(c *orpc.Context) (any, error) {
		in, _ := orpc.Bind[listDomainsInput](c)

		doms, err := db.ListAllDomains(c.Context)
		if err != nil {
			return nil, err
		}

		dtos := make([]domainDTO, 0, len(doms))
		for _, dom := range doms {
			if in != nil && in.ProjectName != "" && in.ServiceName != "" {
				if dom.ProjectName != in.ProjectName || dom.ServiceName != in.ServiceName {
					continue
				}
			}
			dtos = append(dtos, toDomainDTO(dom))
		}
		return dtos, nil
	})

	d.Register("domains/createDomain", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}

		in, err := orpc.Bind[createDomainInput](c)
		if err != nil {
			return nil, err
		}

		host := strings.TrimSpace(in.Host)
		if host == "" {
			host = strings.TrimSpace(in.DomainName)
		}
		if host == "" {
			return nil, orpc.NewBadRequest("host cannot be empty")
		}

		proj, err := db.GetProjectByName(c.Context, in.ProjectName)
		if err != nil {
			return nil, domain.ErrNotFound
		}

		svc, err := db.GetServiceByName(c.Context, proj.ID, in.ServiceName)
		if err != nil {
			return nil, domain.ErrNotFound
		}

		portNum := in.Port
		if portNum <= 0 {
			portNum = 80
		}
		path := in.Path
		if path == "" {
			path = "/"
		}

		dom := &domain.Domain{
			ID:          domain.NewID(),
			ServiceID:   svc.ID,
			DomainName:  host,
			Path:        path,
			Port:        portNum,
			HTTPS:       in.HTTPS,
			CertMode:    in.CertificateResolver,
			Middlewares: in.Middlewares,
			ProjectName: proj.Name,
			ServiceName: svc.Name,
			Status:      "active",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := db.CreateDomain(c.Context, dom); err != nil {
			return nil, err
		}

		if proxy != nil {
			_ = proxy.ApplyRoute(c.Context, dom.ToRouteConfig())
		}

		return toDomainDTO(dom), nil
	})

	d.Register("domains/updateDomain", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}

		in, err := orpc.Bind[updateDomainInput](c)
		if err != nil {
			return nil, err
		}
		if in.ID == "" {
			return nil, orpc.NewBadRequest("domain id is required")
		}

		dom, err := db.GetDomain(c.Context, in.ID)
		if err != nil {
			return nil, domain.ErrNotFound
		}

		host := strings.TrimSpace(in.Host)
		if host == "" {
			host = strings.TrimSpace(in.DomainName)
		}
		if host != "" {
			dom.DomainName = host
		}
		if in.Port > 0 {
			dom.Port = in.Port
		}
		if in.Path != "" {
			dom.Path = in.Path
		}
		dom.HTTPS = in.HTTPS
		if in.CertificateResolver != "" {
			dom.CertMode = in.CertificateResolver
		}
		if in.Middlewares != nil {
			dom.Middlewares = in.Middlewares
		}
		dom.UpdatedAt = time.Now().UTC()

		if err := db.UpdateDomain(c.Context, dom); err != nil {
			return nil, err
		}

		if proxy != nil {
			_ = proxy.ApplyRoute(c.Context, dom.ToRouteConfig())
		}

		return toDomainDTO(dom), nil
	})

	d.Register("domains/deleteDomain", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}

		in, err := orpc.Bind[deleteDomainInput](c)
		if err != nil {
			return nil, err
		}

		if in.ID != "" {
			if dom, err := db.GetDomain(c.Context, in.ID); err == nil {
				_ = db.DeleteDomain(c.Context, in.ID)
				if proxy != nil {
					_ = proxy.RemoveRoute(c.Context, dom.ServiceID)
				}
			}
			return map[string]bool{"success": true}, nil
		}

		if in.Host != "" {
			doms, _ := db.ListAllDomains(c.Context)
			for _, dom := range doms {
				if dom.DomainName == in.Host {
					_ = db.DeleteDomain(c.Context, dom.ID)
					if proxy != nil {
						_ = proxy.RemoveRoute(c.Context, dom.ServiceID)
					}
					break
				}
			}
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("domains/getPrimaryDomain", func(c *orpc.Context) (any, error) {
		val, _ := db.GetSetting(c.Context, "primary_domain")
		return map[string]any{"domain": val, "id": val}, nil
	})

	d.Register("domains/setPrimaryDomain", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}

		in, err := orpc.Bind[setPrimaryDomainInput](c)
		if err != nil {
			return nil, err
		}

		target := in.Domain
		if target == "" {
			target = in.ID
		}
		_ = db.SetSetting(c.Context, "primary_domain", target)
		return map[string]bool{"success": true}, nil
	})
}
