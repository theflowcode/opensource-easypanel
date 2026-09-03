package domain

// DatabaseConfig stores database-specific settings, remote exposure, and 1-click GUI tools.
type DatabaseConfig struct {
	DatabaseName string   `json:"databaseName,omitempty"`
	RootPassword string   `json:"rootPassword,omitempty"`
	ExposePort   int      `json:"exposePort,omitempty"`   // External host port if exposed (e.g. 5432, 6379)
	IsExposed    bool     `json:"isExposed,omitempty"`
	EnabledTools []string `json:"enabledTools,omitempty"` // e.g. ["pgweb", "dbgate", "redis-commander"]
	InternalURL  string   `json:"internalUrl,omitempty"`  // In-network connection string
	ExternalURL  string   `json:"externalUrl,omitempty"`  // Remote host connection string
}

// IsDatabase returns true if the service type represents a database engine.
func (t ServiceType) IsDatabase() bool {
	switch t {
	case ServiceTypeDatabase, ServiceTypePostgres, ServiceTypeRedis,
		ServiceTypeMySQL, ServiceTypeMariaDB, ServiceTypeMongoDB:
		return true
	default:
		return false
	}
}
