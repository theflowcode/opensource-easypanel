package http

import (
	"fmt"
	"strings"


	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type dbCredentialsInput struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
	Password    string `json:"password"`
}

// registerServicesDBRoutes binds database service engines to the oRPC dispatcher.
func registerServicesDBRoutes(d *orpc.Dispatcher, db port.DatabasePort, docker port.DockerPort) {
	engines := []string{"postgres", "redis", "mysql", "mariadb", "mongo"}

	for _, eng := range engines {
		engine := eng // closure capture

		d.Register(fmt.Sprintf("services/%s/inspectService", engine), func(c *orpc.Context) (any, error) {
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

			dbConfig := svc.DatabaseConfig
			if dbConfig == nil {
				dbConfig = &domain.DatabaseConfig{
					ExposePort:   0,
					IsExposed:    false,
					EnabledTools: []string{},
				}
			}

			envStr := strings.Join(domain.EnvVarsToSlice(svc.EnvVars), "\n")

			return map[string]any{
				"id":          svc.ID,
				"projectName": proj.Name,
				"name":        svc.Name,
				"type":        engine,
				"image":       svc.Image,
				"env":         envStr,
				"enabled":     svc.Status != domain.ServiceStatusStopped,
				"database":    dbConfig,
				"resources": map[string]any{
					"memoryReservation": svc.Resources.MemoryLimit,
					"memoryLimit":       svc.Resources.MemoryLimit,
					"cpuReservation":    svc.Resources.CPULimit,
					"cpuLimit":          svc.Resources.CPULimit,
				},
				"createdAt": svc.CreatedAt,
				"updatedAt": svc.UpdatedAt,
			}, nil
		})

		d.Register(fmt.Sprintf("services/%s/enableService", engine), func(c *orpc.Context) (any, error) {
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

		d.Register(fmt.Sprintf("services/%s/disableService", engine), func(c *orpc.Context) (any, error) {
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
			svc.Status = domain.ServiceStatusStopped
			_ = db.UpdateService(c.Context, svc)
			return map[string]bool{"success": true}, nil
		})

		d.Register(fmt.Sprintf("services/%s/updateCredentials", engine), func(c *orpc.Context) (any, error) {
			if err := c.RequireAdmin(); err != nil {
				return nil, err
			}
			in, err := orpc.Bind[dbCredentialsInput](c)
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
			if svc.DatabaseConfig == nil {
				svc.DatabaseConfig = &domain.DatabaseConfig{}
			}
			svc.DatabaseConfig.RootPassword = in.Password
			_ = db.UpdateService(c.Context, svc)
			return map[string]bool{"success": true}, nil
		})

		d.Register(fmt.Sprintf("services/%s/updateResources", engine), func(c *orpc.Context) (any, error) {
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
			_ = db.UpdateService(c.Context, svc)
			return map[string]bool{"success": true}, nil
		})

		d.Register(fmt.Sprintf("services/%s/destroyService", engine), func(c *orpc.Context) (any, error) {
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
			_ = db.DeleteService(c.Context, svc.ID)
			return map[string]bool{"success": true}, nil
		})

		d.Register(fmt.Sprintf("services/%s/exposeService", engine), func(c *orpc.Context) (any, error) {
			if err := c.RequireAdmin(); err != nil {
				return nil, err
			}
			return map[string]bool{"success": true}, nil
		})

		d.Register(fmt.Sprintf("services/%s/updateAdvanced", engine), func(c *orpc.Context) (any, error) {
			if err := c.RequireAdmin(); err != nil {
				return nil, err
			}
			return map[string]bool{"success": true}, nil
		})

		d.Register(fmt.Sprintf("services/%s/enableDbGate", engine), func(c *orpc.Context) (any, error) {
			if err := c.RequireAdmin(); err != nil {
				return nil, err
			}
			return map[string]bool{"success": true}, nil
		})
		d.Register(fmt.Sprintf("services/%s/disableDbGate", engine), func(c *orpc.Context) (any, error) {
			if err := c.RequireAdmin(); err != nil {
				return nil, err
			}
			return map[string]bool{"success": true}, nil
		})

		switch engine {
		case "postgres":
			d.Register("services/postgres/enablePgWeb", func(c *orpc.Context) (any, error) {
				if err := c.RequireAdmin(); err != nil {
					return nil, err
				}
				return map[string]bool{"success": true}, nil
			})
			d.Register("services/postgres/disablePgWeb", func(c *orpc.Context) (any, error) {
				if err := c.RequireAdmin(); err != nil {
					return nil, err
				}
				return map[string]bool{"success": true}, nil
			})
		case "mysql", "mariadb":
			d.Register(fmt.Sprintf("services/%s/enablePhpMyAdmin", engine), func(c *orpc.Context) (any, error) {
				if err := c.RequireAdmin(); err != nil {
					return nil, err
				}
				return map[string]bool{"success": true}, nil
			})
			d.Register(fmt.Sprintf("services/%s/disablePhpMyAdmin", engine), func(c *orpc.Context) (any, error) {
				if err := c.RequireAdmin(); err != nil {
					return nil, err
				}
				return map[string]bool{"success": true}, nil
			})
		case "redis":
			d.Register("services/redis/enableRedisCommander", func(c *orpc.Context) (any, error) {
				if err := c.RequireAdmin(); err != nil {
					return nil, err
				}
				return map[string]bool{"success": true}, nil
			})
			d.Register("services/redis/disableRedisCommander", func(c *orpc.Context) (any, error) {
				if err := c.RequireAdmin(); err != nil {
					return nil, err
				}
				return map[string]bool{"success": true}, nil
			})
		case "mongo":
			d.Register("services/mongo/enableMongoExpress", func(c *orpc.Context) (any, error) {
				if err := c.RequireAdmin(); err != nil {
					return nil, err
				}
				return map[string]bool{"success": true}, nil
			})
			d.Register("services/mongo/disableMongoExpress", func(c *orpc.Context) (any, error) {
				if err := c.RequireAdmin(); err != nil {
					return nil, err
				}
				return map[string]bool{"success": true}, nil
			})
		}
	}
}
