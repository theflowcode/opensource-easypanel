package http

import (
	"strings"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type updateSourceGitInput struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
	RepoURL     string `json:"repoUrl"`
	Branch      string `json:"branch"`
}

type updateSourceDockerfileInput struct {
	ProjectName    string `json:"projectName"`
	ServiceName    string `json:"serviceName"`
	DockerfilePath string `json:"dockerfilePath"`
	ContextPath    string `json:"contextPath"`
}

// registerServicesAppConfigRoutes binds app configuration procedures to the oRPC dispatcher.
func registerServicesAppConfigRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	d.Register("services/app/updateBasicAuth", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/updateScripts", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/updateRedirects", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/updateMaintenance", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/updateBuild", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/enableGithubDeploy", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/disableGithubDeploy", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/updateSourceGit", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[updateSourceGitInput](c)
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
		svc.SourceType = domain.SourceTypeGit
		if svc.SourceConfig == nil {
			svc.SourceConfig = &domain.SourceConfig{}
		}
		svc.SourceConfig.RepoURL = strings.TrimSpace(in.RepoURL)
		svc.SourceConfig.Branch = strings.TrimSpace(in.Branch)
		if err := db.UpdateService(c.Context, svc); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/updateSourceDockerfile", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[updateSourceDockerfileInput](c)
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
		svc.SourceType = domain.SourceTypeDockerfile
		if svc.SourceConfig == nil {
			svc.SourceConfig = &domain.SourceConfig{}
		}
		svc.SourceConfig.DockerfilePath = strings.TrimSpace(in.DockerfilePath)
		svc.SourceConfig.ContextPath = strings.TrimSpace(in.ContextPath)
		if err := db.UpdateService(c.Context, svc); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/startService", func(c *orpc.Context) (any, error) {
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
		svc.Status = domain.ServiceStatusRunning
		_ = db.UpdateService(c.Context, svc)
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/refreshDeployToken", func(c *orpc.Context) (any, error) {
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
		svc.DeployToken = domain.NewID()
		_ = db.UpdateService(c.Context, svc)
		return map[string]string{"deployToken": svc.DeployToken}, nil
	})

	d.Register("services/app/getExposedPorts", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return []int{}, nil
	})
}
