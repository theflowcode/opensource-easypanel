package http

import (
	"time"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

// registerServicesAppOpsRoutes binds lifecycle operations for app services.
func registerServicesAppOpsRoutes(d *orpc.Dispatcher, db port.DatabasePort, docker port.DockerPort) {
	d.Register("services/app/deployService", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
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
		if docker != nil {
			_, _ = docker.DeployService(c.Context, svc.ToSpec())
		}
		_ = db.CreateAction(c.Context, &domain.Action{
			ID:          domain.NewID(),
			ProjectName: proj.Name,
			ServiceName: svc.Name,
			Type:        domain.ActionTypeDeployment,
			Status:      domain.ActionStatusDone,
			Description: "Deployed service " + svc.Name,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		})
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/restartService", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
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
		if docker != nil {
			_ = docker.RestartService(c.Context, svc.ID)
		}
		_ = db.CreateAction(c.Context, &domain.Action{
			ID:          domain.NewID(),
			ProjectName: proj.Name,
			ServiceName: svc.Name,
			Type:        domain.ActionTypeServiceRestart,
			Status:      domain.ActionStatusDone,
			Description: "Restarted service " + svc.Name,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		})
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/stopService", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
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
		if docker != nil {
			_ = docker.StopService(c.Context, svc.ID)
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/destroyService", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
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
		if docker != nil {
			_ = docker.DeleteService(c.Context, svc.ID)
		}
		if err := db.DeleteService(c.Context, svc.ID); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})
}
