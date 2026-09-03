package mock

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.TemplatePort = (*MockTemplatePort)(nil)

// MockTemplatePort is a thread-safe mock implementing port.TemplatePort.
type MockTemplatePort struct {
	mu sync.RWMutex

	// Customizable function hooks
	ListTemplatesFunc  func(ctx context.Context) ([]domain.TemplateSummary, error)
	GetTemplateFunc    func(ctx context.Context, id string) (*domain.Template, error)
	ParseTemplateFunc  func(ctx context.Context, rawContent []byte) (*domain.Template, error)
	RenderTemplateFunc func(ctx context.Context, tmpl *domain.Template, values map[string]string) ([]domain.ServiceSpec, error)

	// In-memory catalog fixture
	Templates map[string]*domain.Template

	// Call tracking
	Calls []string
}

func NewMockTemplatePort() *MockTemplatePort {
	m := &MockTemplatePort{
		Templates: make(map[string]*domain.Template),
	}
	// Seed with default fixture
	m.Templates["postgres"] = &domain.Template{
		ID:          "postgres",
		Name:        "PostgreSQL",
		Description: "Object-relational database system",
		Category:    "database",
		Icon:        "postgres",
		Variables: []domain.TemplateVariable{
			{Key: "POSTGRES_PASSWORD", Label: "Password", Required: true, IsSecret: true, DefaultValue: "postgres"},
		},
		Services: []domain.TemplateServiceSpec{
			{
				Name:  "postgres",
				Image: "postgres:16-alpine",
				Ports: []domain.PortMapping{
					{HostPort: 5432, ContainerPort: 5432, Protocol: "tcp"},
				},
				EnvVars: []domain.EnvVar{
					{Key: "POSTGRES_PASSWORD", Value: "${POSTGRES_PASSWORD}"},
				},
			},
		},
	}
	return m
}

// Reset clears all in-memory mock state and recorded calls.
func (m *MockTemplatePort) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Templates = make(map[string]*domain.Template)
	m.Calls = nil
	m.ListTemplatesFunc = nil
	m.GetTemplateFunc = nil
	m.ParseTemplateFunc = nil
	m.RenderTemplateFunc = nil
}

func (m *MockTemplatePort) ListTemplates(ctx context.Context) ([]domain.TemplateSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ListTemplatesFunc != nil {
		return m.ListTemplatesFunc(ctx)
	}

	var list []domain.TemplateSummary
	for _, t := range m.Templates {
		list = append(list, domain.TemplateSummary{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Category:    t.Category,
			Icon:        t.Icon,
		})
	}
	return list, nil
}

func (m *MockTemplatePort) GetTemplate(ctx context.Context, id string) (*domain.Template, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetTemplateFunc != nil {
		return m.GetTemplateFunc(ctx, id)
	}

	tmpl, ok := m.Templates[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	clone := *tmpl
	return &clone, nil
}

func (m *MockTemplatePort) ParseTemplate(ctx context.Context, rawContent []byte) (*domain.Template, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "ParseTemplate")

	if m.ParseTemplateFunc != nil {
		return m.ParseTemplateFunc(ctx, rawContent)
	}

	var tmpl domain.Template
	if err := json.Unmarshal(rawContent, &tmpl); err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (m *MockTemplatePort) RenderTemplate(ctx context.Context, tmpl *domain.Template, values map[string]string) ([]domain.ServiceSpec, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "RenderTemplate")

	if m.RenderTemplateFunc != nil {
		return m.RenderTemplateFunc(ctx, tmpl, values)
	}

	var specs []domain.ServiceSpec
	for _, s := range tmpl.Services {
		var envs []domain.EnvVar
		for _, e := range s.EnvVars {
			val := e.Value
			if v, ok := values[e.Key]; ok {
				val = v
			}
			envs = append(envs, domain.EnvVar{Key: e.Key, Value: val})
		}

		specs = append(specs, domain.ServiceSpec{
			ID:        tmpl.ID + "-" + s.Name,
			Name:      s.Name,
			Type:      domain.ServiceTypeDatabase,
			Image:     s.Image,
			Ports:     s.Ports,
			Volumes:   s.Volumes,
			EnvVars:   envs,
			Replicas:  1,
			Resources: s.Resources,
		})
	}
	return specs, nil
}
