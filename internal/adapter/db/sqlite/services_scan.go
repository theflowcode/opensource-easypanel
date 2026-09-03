package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (r *Repository) scanService(row *sql.Row) (*domain.Service, error) {
	return r.scanServiceRow(row)
}

func (r *Repository) scanServiceRow(scanner rowScanner) (*domain.Service, error) {
	var (
		s             domain.Service
		srvType       string
		sourceType    string
		sourceCfgJSON string
		healthJSON    string
		cronJobsJSON  string
		dbCfgJSON     string
		redirectsJSON string
		zdInt         int
		labelsJSON    string
		status        string
		argsJSON      string
		envVarsJSON   string
		portsJSON     string
		volumesJSON   string
		domainsJSON   string
	)

	err := scanner.Scan(
		&s.ID, &s.ProjectID, &s.ProjectName, &s.Name, &srvType, &s.DeployToken, &s.DeployScript,
		&sourceType, &sourceCfgJSON, &s.Image, &s.Command, &argsJSON,
		&envVarsJSON, &portsJSON, &volumesJSON, &domainsJSON, &s.Replicas,
		&s.Resources.CPULimit, &s.Resources.MemoryLimit, &s.RestartPolicy, &healthJSON, &cronJobsJSON,
		&dbCfgJSON, &redirectsJSON, &s.PrimaryDomainID, &zdInt, &labelsJSON,
		&status, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	s.Type = domain.ServiceType(srvType)
	s.SourceType = domain.ServiceSourceType(sourceType)
	s.Status = domain.ServiceStatus(status)
	s.ZeroDowntime = zdInt == 1

	_ = json.Unmarshal([]byte(argsJSON), &s.Args)
	_ = json.Unmarshal([]byte(envVarsJSON), &s.EnvVars)
	_ = json.Unmarshal([]byte(portsJSON), &s.Ports)
	_ = json.Unmarshal([]byte(volumesJSON), &s.Volumes)
	_ = json.Unmarshal([]byte(domainsJSON), &s.Domains)
	_ = json.Unmarshal([]byte(sourceCfgJSON), &s.SourceConfig)
	_ = json.Unmarshal([]byte(healthJSON), &s.HealthCheck)
	_ = json.Unmarshal([]byte(cronJobsJSON), &s.CronJobs)
	_ = json.Unmarshal([]byte(dbCfgJSON), &s.DatabaseConfig)
	_ = json.Unmarshal([]byte(redirectsJSON), &s.Redirects)
	_ = json.Unmarshal([]byte(labelsJSON), &s.Labels)

	return &s, nil
}
