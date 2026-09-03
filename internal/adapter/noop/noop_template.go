package noop

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.TemplatePort = (*NoOpTemplate)(nil)

// NoOpTemplate is a Null Object implementation of port.TemplatePort when template catalogs are disabled.
type NoOpTemplate struct{}

func NewNoOpTemplate() *NoOpTemplate {
	return &NoOpTemplate{}
}

func (n *NoOpTemplate) ListTemplates(ctx context.Context) ([]domain.TemplateSummary, error) {
	return []domain.TemplateSummary{}, nil
}

func (n *NoOpTemplate) GetTemplate(ctx context.Context, id string) (*domain.Template, error) {
	return nil, domain.ErrNotFound
}

func (n *NoOpTemplate) ParseTemplate(ctx context.Context, rawContent []byte) (*domain.Template, error) {
	return nil, domain.ErrNotFound
}

func (n *NoOpTemplate) RenderTemplate(ctx context.Context, tmpl *domain.Template, values map[string]string) ([]domain.ServiceSpec, error) {
	return []domain.ServiceSpec{}, nil
}
