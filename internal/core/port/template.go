package port

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// TemplatePort defines contracts for parsing, validating, and instantiating 1-click templates.
type TemplatePort interface {
	// ListTemplates returns lightweight summaries of available 1-click templates.
	ListTemplates(ctx context.Context) ([]domain.TemplateSummary, error)

	// GetTemplate retrieves the full template specification by ID.
	GetTemplate(ctx context.Context, id string) (*domain.Template, error)

	// ParseTemplate parses raw Compose or Easypanel schema bytes into a Template model.
	ParseTemplate(ctx context.Context, rawContent []byte) (*domain.Template, error)

	// RenderTemplate interpolates user-provided variables into actionable ServiceSpecs.
	RenderTemplate(ctx context.Context, tmpl *domain.Template, values map[string]string) ([]domain.ServiceSpec, error)
}
