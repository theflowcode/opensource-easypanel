package http

import (
	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type setNotesInput struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
	Notes       string `json:"notes"`
}

type renameServiceInput struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
	NewName     string `json:"newName"`
}

// registerServicesCommonRoutes binds common service utilities and compose handlers to the oRPC dispatcher.
func registerServicesCommonRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	d.Register("services/common/getNotes", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[serviceTargetInput](c)
		if err != nil {
			return nil, err
		}
		proj, err := db.GetProjectByName(c.Context, in.ProjectName)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		svc, err := db.GetServiceByName(c.Context, proj.ID, in.ServiceName)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return map[string]string{"notes": svc.Notes}, nil
	})

	d.Register("services/common/setNotes", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[setNotesInput](c)
		if err != nil {
			return nil, err
		}
		proj, err := db.GetProjectByName(c.Context, in.ProjectName)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		svc, err := db.GetServiceByName(c.Context, proj.ID, in.ServiceName)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		svc.Notes = in.Notes
		if err := db.UpdateService(c.Context, svc); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/common/getServiceError", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[serviceTargetInput](c)
		if err != nil {
			return nil, err
		}
		proj, err := db.GetProjectByName(c.Context, in.ProjectName)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		svc, err := db.GetServiceByName(c.Context, proj.ID, in.ServiceName)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		var errVal any
		if svc.LastError != "" {
			errVal = svc.LastError
		}
		return map[string]any{"error": errVal}, nil
	})

	d.Register("services/common/rename", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[renameServiceInput](c)
		if err != nil {
			return nil, err
		}
		proj, err := db.GetProjectByName(c.Context, in.ProjectName)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		svc, err := db.GetServiceByName(c.Context, proj.ID, in.ServiceName)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		svc.Name = in.NewName
		if err := db.UpdateService(c.Context, svc); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	// Compose service procedures
	d.Register("services/compose/inspectService", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return map[string]any{"type": "compose"}, nil
	})
	d.Register("services/compose/deployService", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})
	d.Register("services/compose/restartService", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})
	d.Register("services/compose/stopService", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})
	d.Register("services/compose/destroyService", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})
	d.Register("services/compose/getDockerServices", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return []any{}, nil
	})
	d.Register("services/compose/getIssues", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return []any{}, nil
	})
}
