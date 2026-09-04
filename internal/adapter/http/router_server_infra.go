package http

import (
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type createMiddlewareInput struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type destroyMiddlewareInput struct {
	ID string `json:"id"`
}

// registerServerInfraRoutes binds cluster, certificates, cloudflare, and middlewares.
func registerServerInfraRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	// Cluster
	d.Register("cluster/listNodes", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return []map[string]any{
			{
				"id":           "node-leader",
				"role":         "manager",
				"status":       "ready",
				"availability": "active",
				"managerStatus": map[string]any{
					"leader":       true,
					"reachability": "reachable",
					"addr":         "127.0.0.1:2377",
				},
				"hostname": "easypanel-host",
			},
		}, nil
	})

	d.Register("cluster/addWorkerCommand", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]string{
			"command": "docker swarm join --token SWMTKN-1-dummy-worker-token 127.0.0.1:2377",
		}, nil
	})

	d.Register("cluster/removeNode", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	// Certificates
	d.Register("certificates/listCertificates", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return []any{}, nil
	})

	d.Register("certificates/removeCertificate", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	// Cloudflare Tunnel
	d.Register("cloudflareTunnel/getConfig", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return map[string]any{
			"enabled":  false,
			"token":    "",
			"tunnelId": "",
		}, nil
	})

	d.Register("cloudflareTunnel/listTunnels", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return []any{}, nil
	})

	d.Register("cloudflareTunnel/listZones", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return []any{}, nil
	})

	// Middlewares
	d.Register("middlewares/listMiddlewares", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		mws, err := db.ListMiddlewares(c.Context)
		if err != nil {
			return nil, err
		}
		if mws == nil {
			return []any{}, nil
		}
		return mws, nil
	})

	d.Register("middlewares/createMiddleware", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[createMiddlewareInput](c)
		if err != nil {
			return nil, err
		}
		mw := &domain.Middleware{
			ID:        domain.NewID(),
			Name:      strings.TrimSpace(in.Name),
			Type:      strings.TrimSpace(in.Type),
			Config:    in.Config,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := db.CreateMiddleware(c.Context, mw); err != nil {
			return nil, err
		}
		return mw, nil
	})

	d.Register("middlewares/destroyMiddleware", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[destroyMiddlewareInput](c)
		if err != nil {
			return nil, err
		}
		if err := db.DeleteMiddleware(c.Context, in.ID); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	registerServerStorageRoutes(d, db)
}
