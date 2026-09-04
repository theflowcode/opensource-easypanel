package http

import (
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type createProjectInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type projectByNameInput struct {
	ProjectName string `json:"projectName"`
	Name        string `json:"name"`
}

type updateProjectEnvInput struct {
	ProjectName string `json:"projectName"`
	Env         string `json:"env"`
}

// registerProjectsRoutes binds project management procedures to the oRPC dispatcher.
func registerProjectsRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	d.Register("projects/canCreateProject", func(c *orpc.Context) (any, error) {
		return true, nil
	})

	d.Register("projects/listProjectsAndServices", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		projects, err := db.ListProjects(c.Context)
		if err != nil {
			return nil, err
		}
		allServices, err := db.ListAllServices(c.Context)
		if err != nil {
			return nil, err
		}

		projMap := make(map[string]string)
		for _, p := range projects {
			projMap[p.ID] = p.Name
		}

		type projDTO struct {
			ID          string    `json:"id"`
			Name        string    `json:"name"`
			Description string    `json:"description"`
			CreatedAt   time.Time `json:"createdAt"`
			UpdatedAt   time.Time `json:"updatedAt"`
		}

		type svcDTO struct {
			ID          string `json:"id"`
			ProjectName string `json:"projectName"`
			Name        string `json:"name"`
			Type        string `json:"type"`
			Image       string `json:"image"`
			Replicas    int    `json:"replicas"`
		}

		pDTOs := make([]projDTO, 0, len(projects))
		for _, p := range projects {
			pDTOs = append(pDTOs, projDTO{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				CreatedAt:   p.CreatedAt,
				UpdatedAt:   p.UpdatedAt,
			})
		}

		sDTOs := make([]svcDTO, 0, len(allServices))
		for _, s := range allServices {
			pName := s.ProjectName
			if pName == "" {
				pName = projMap[s.ProjectID]
			}
			sType := string(s.Type)
			if sType == "" {
				sType = "app"
			}
			sDTOs = append(sDTOs, svcDTO{
				ID:          s.ID,
				ProjectName: pName,
				Name:        s.Name,
				Type:        sType,
				Image:       s.Image,
				Replicas:    s.Replicas,
			})
		}

		return map[string]any{
			"projects": pDTOs,
			"services": sDTOs,
		}, nil
	})

	d.Register("projects/listProjects", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		projects, err := db.ListProjects(c.Context)
		if err != nil {
			return nil, err
		}
		return projects, nil
	})

	d.Register("projects/createProject", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}

		in, err := orpc.Bind[createProjectInput](c)
		if err != nil {
			return nil, err
		}

		name := strings.TrimSpace(in.Name)
		if name == "" {
			return nil, orpc.NewBadRequest("project name cannot be empty")
		}

		proj := &domain.Project{
			ID:          domain.NewID(),
			Name:        name,
			Description: in.Description,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := db.CreateProject(c.Context, proj); err != nil {
			return nil, err
		}

		return map[string]any{
			"id":          proj.ID,
			"name":        proj.Name,
			"description": proj.Description,
			"createdAt":   proj.CreatedAt,
			"updatedAt":   proj.UpdatedAt,
		}, nil
	})

	d.Register("projects/destroyProject", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}

		in, err := orpc.Bind[projectByNameInput](c)
		if err != nil {
			return nil, err
		}

		name := in.ProjectName
		if name == "" {
			name = in.Name
		}
		if name == "" {
			return nil, orpc.NewBadRequest("project name is required")
		}

		proj, err := db.GetProjectByName(c.Context, name)
		if err != nil {
			return nil, domain.ErrNotFound
		}

		if err := db.DeleteProject(c.Context, proj.ID); err != nil {
			return nil, err
		}

		return map[string]bool{"success": true}, nil
	})

	d.Register("projects/inspectProject", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[projectByNameInput](c)
		if err != nil {
			return nil, err
		}

		name := in.ProjectName
		if name == "" {
			name = in.Name
		}
		if name == "" {
			return nil, orpc.NewBadRequest("project name is required")
		}

		proj, err := db.GetProjectByName(c.Context, name)
		if err != nil {
			return nil, domain.ErrNotFound
		}

		services, err := db.ListServicesByProject(c.Context, proj.ID)
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"id":          proj.ID,
			"name":        proj.Name,
			"description": proj.Description,
			"env":         proj.Env,
			"services":    services,
			"createdAt":   proj.CreatedAt,
			"updatedAt":   proj.UpdatedAt,
		}, nil
	})

	d.Register("projects/updateProjectEnv", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}

		in, err := orpc.Bind[updateProjectEnvInput](c)
		if err != nil {
			return nil, err
		}

		proj, err := db.GetProjectByName(c.Context, in.ProjectName)
		if err != nil {
			return nil, domain.ErrNotFound
		}

		proj.Env = in.Env
		if err := db.UpdateProject(c.Context, proj); err != nil {
			return nil, err
		}

		return map[string]bool{"success": true}, nil
	})

	d.Register("projects/getDockerContainers", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return []any{}, nil
	})
}
