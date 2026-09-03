package domain

// Template represents a 1-click app or database schema template.
type Template struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Category    string                `json:"category"`
	Icon        string                `json:"icon"`
	Variables   []TemplateVariable    `json:"variables"`
	Services    []TemplateServiceSpec `json:"services"`
}

// TemplateSummary is a lightweight template description for catalog display.
type TemplateSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Icon        string `json:"icon"`
}

// TemplateVariable defines configurable inputs for 1-click templates.
type TemplateVariable struct {
	Key          string `json:"key"`
	Label        string `json:"label"`
	Description  string `json:"description"`
	DefaultValue string `json:"defaultValue"`
	Required     bool   `json:"required"`
	IsSecret     bool   `json:"isSecret"`
}

// TemplateServiceSpec represents a service within a template.
type TemplateServiceSpec struct {
	Name      string         `json:"name"`
	Image     string         `json:"image"`
	Ports     []PortMapping  `json:"ports,omitempty"`
	Volumes   []VolumeMount  `json:"volumes,omitempty"`
	EnvVars   []EnvVar       `json:"envVars,omitempty"`
	Resources ResourceLimits `json:"resources"`
}
