package http

import (
	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type portInput struct {
	ProjectName   string `json:"projectName"`
	ServiceName   string `json:"serviceName"`
	HostPort      int    `json:"hostPort"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

type deletePortInput struct {
	ProjectName   string `json:"projectName"`
	ServiceName   string `json:"serviceName"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

// registerPortsRoutes binds container port exposure procedures to the oRPC dispatcher.
func registerPortsRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	d.Register("ports/listPorts", func(c *orpc.Context) (any, error) {
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
		if svc.Ports == nil {
			return []domain.PortMapping{}, nil
		}
		return svc.Ports, nil
	})

	d.Register("ports/createPort", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}

		in, err := orpc.Bind[portInput](c)
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

		proto := in.Protocol
		if proto == "" {
			proto = "tcp"
		}
		mapping := domain.PortMapping{
			HostPort:      in.HostPort,
			ContainerPort: in.ContainerPort,
			Protocol:      proto,
		}

		svc.Ports = append(svc.Ports, mapping)
		if err := db.UpdateService(c.Context, svc); err != nil {
			return nil, err
		}
		return mapping, nil
	})

	d.Register("ports/updatePort", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[portInput](c)
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

		for i, p := range svc.Ports {
			if p.ContainerPort == in.ContainerPort {
				svc.Ports[i].HostPort = in.HostPort
				if in.Protocol != "" {
					svc.Ports[i].Protocol = in.Protocol
				}
				break
			}
		}
		_ = db.UpdateService(c.Context, svc)
		return map[string]bool{"success": true}, nil
	})

	d.Register("ports/deletePort", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[deletePortInput](c)
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

		filtered := make([]domain.PortMapping, 0, len(svc.Ports))
		for _, p := range svc.Ports {
			if p.ContainerPort != in.ContainerPort {
				filtered = append(filtered, p)
			}
		}
		svc.Ports = filtered
		_ = db.UpdateService(c.Context, svc)
		return map[string]bool{"success": true}, nil
	})

	d.Register("ports/deleteAllPorts", func(c *orpc.Context) (any, error) {
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
		svc.Ports = []domain.PortMapping{}
		_ = db.UpdateService(c.Context, svc)
		return map[string]bool{"success": true}, nil
	})
}
