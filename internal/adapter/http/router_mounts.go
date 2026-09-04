package http

import (
	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type mountInput struct {
	ProjectName   string `json:"projectName"`
	ServiceName   string `json:"serviceName"`
	Type          string `json:"type"` // "volume", "bind", or "file"
	Name          string `json:"name"`
	HostPath      string `json:"hostPath"`
	ContainerPath string `json:"containerPath"`
	Content       string `json:"content"`
	ReadOnly      bool   `json:"readOnly"`
}

type deleteMountInput struct {
	ProjectName   string `json:"projectName"`
	ServiceName   string `json:"serviceName"`
	Name          string `json:"name"`
	ContainerPath string `json:"containerPath"`
}

// registerMountsRoutes binds storage mounts procedures to the oRPC dispatcher.
func registerMountsRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	d.Register("mounts/listMounts", func(c *orpc.Context) (any, error) {
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
		if svc.Volumes == nil {
			return []domain.VolumeMount{}, nil
		}
		return svc.Volumes, nil
	})

	d.Register("mounts/createMount", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}

		in, err := orpc.Bind[mountInput](c)
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

		mType := in.Type
		if mType == "" {
			mType = "volume"
		}

		mount := domain.VolumeMount{
			Type:          mType,
			Name:          in.Name,
			HostPath:      in.HostPath,
			ContainerPath: in.ContainerPath,
			Content:       in.Content,
			ReadOnly:      in.ReadOnly,
		}

		svc.Volumes = append(svc.Volumes, mount)
		if err := db.UpdateService(c.Context, svc); err != nil {
			return nil, err
		}
		return mount, nil
	})

	d.Register("mounts/updateMount", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[mountInput](c)
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

		for i, v := range svc.Volumes {
			if v.Name == in.Name || (in.ContainerPath != "" && v.ContainerPath == in.ContainerPath) {
				if in.Type != "" {
					svc.Volumes[i].Type = in.Type
				}
				if in.HostPath != "" {
					svc.Volumes[i].HostPath = in.HostPath
				}
				if in.ContainerPath != "" {
					svc.Volumes[i].ContainerPath = in.ContainerPath
				}
				svc.Volumes[i].Content = in.Content
				svc.Volumes[i].ReadOnly = in.ReadOnly
				break
			}
		}
		_ = db.UpdateService(c.Context, svc)
		return map[string]bool{"success": true}, nil
	})

	d.Register("mounts/deleteMount", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[deleteMountInput](c)
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

		filtered := make([]domain.VolumeMount, 0, len(svc.Volumes))
		for _, v := range svc.Volumes {
			if v.Name != in.Name && (in.ContainerPath == "" || v.ContainerPath != in.ContainerPath) {
				filtered = append(filtered, v)
			}
		}
		svc.Volumes = filtered
		_ = db.UpdateService(c.Context, svc)
		return map[string]bool{"success": true}, nil
	})
}
