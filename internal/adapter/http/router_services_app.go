package http

import (
	"strings"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type serviceTargetInput struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
}

type updateServiceEnvInput struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
	Env         string `json:"env"`
}

type updateServiceDeployInput struct {
	ProjectName  string `json:"projectName"`
	ServiceName  string `json:"serviceName"`
	Replicas     int    `json:"replicas"`
	ZeroDowntime bool   `json:"zeroDowntime"`
}

type updateServiceResourcesInput struct {
	ProjectName string  `json:"projectName"`
	ServiceName string  `json:"serviceName"`
	MemoryLimit int     `json:"memoryLimit"`
	CPULimit    float64 `json:"cpuLimit"`
}

type updateSourceImageInput struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
	Image       string `json:"image"`
}

// registerServicesAppRoutes binds app service procedures to the oRPC dispatcher.
func registerServicesAppRoutes(d *orpc.Dispatcher, db port.DatabasePort, docker port.DockerPort) {
	d.Register("services/app/inspectService", func(c *orpc.Context) (any, error) {
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

		sType := string(svc.Type)
		if sType == "" {
			sType = "app"
		}

		envStr := strings.Join(domain.EnvVarsToSlice(svc.EnvVars), "\n")
		deployUrl := "/api/deploy?token=" + svc.DeployToken

		return map[string]any{
			"id":            svc.ID,
			"projectName":   proj.Name,
			"name":          svc.Name,
			"type":          sType,
			"image":         svc.Image,
			"env":           envStr,
			"enabled":       svc.Status != domain.ServiceStatusStopped,
			"deploymentUrl": deployUrl,
			"deployToken":   svc.DeployToken,
			"source": map[string]any{
				"type":  string(svc.SourceType),
				"image": svc.Image,
			},
			"deploy": map[string]any{
				"replicas":     svc.Replicas,
				"zeroDowntime": svc.ZeroDowntime,
			},
			"resources": map[string]any{
				"memoryReservation": svc.Resources.MemoryLimit,
				"memoryLimit":       svc.Resources.MemoryLimit,
				"cpuReservation":    svc.Resources.CPULimit,
				"cpuLimit":          svc.Resources.CPULimit,
			},
			"ports":       svc.Ports,
			"mounts":      svc.Volumes,
			"redirects":   svc.Redirects,
			"basicAuth":   []any{},
			"scripts":     []any{},
			"maintenance": map[string]any{"enabled": false, "customCss": "", "customLogo": ""},
			"createdAt":   svc.CreatedAt,
			"updatedAt":   svc.UpdatedAt,
		}, nil
	})

	d.Register("services/app/updateEnv", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[updateServiceEnvInput](c)
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
		svc.EnvVars = domain.EnvVarsFromSlice(strings.Split(in.Env, "\n"))
		if err := db.UpdateService(c.Context, svc); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/updateDeploy", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[updateServiceDeployInput](c)
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
		if in.Replicas > 0 {
			svc.Replicas = in.Replicas
		}
		svc.ZeroDowntime = in.ZeroDowntime
		if err := db.UpdateService(c.Context, svc); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/updateResources", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[updateServiceResourcesInput](c)
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
		svc.Resources.MemoryLimit = int64(in.MemoryLimit)
		svc.Resources.CPULimit = in.CPULimit
		if err := db.UpdateService(c.Context, svc); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("services/app/updateSourceImage", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[updateSourceImageInput](c)
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
		svc.Image = strings.TrimSpace(in.Image)
		svc.SourceType = domain.SourceTypeImage
		if err := db.UpdateService(c.Context, svc); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	registerServicesAppOpsRoutes(d, db, docker)
}
