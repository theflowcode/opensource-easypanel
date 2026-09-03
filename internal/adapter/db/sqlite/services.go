package sqlite

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (r *Repository) CreateService(ctx context.Context, s *domain.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := s.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}

	argsJSON, _ := json.Marshal(s.Args)
	envVarsJSON, _ := json.Marshal(s.EnvVars)
	portsJSON, _ := json.Marshal(s.Ports)
	volumesJSON, _ := json.Marshal(s.Volumes)
	domainsJSON, _ := json.Marshal(s.Domains)
	sourceCfgJSON, _ := json.Marshal(s.SourceConfig)
	healthJSON, _ := json.Marshal(s.HealthCheck)
	cronJobsJSON, _ := json.Marshal(s.CronJobs)
	labelsJSON, _ := json.Marshal(s.Labels)

	replicas := s.Replicas
	if replicas <= 0 {
		replicas = 1
	}
	restartPolicy := s.RestartPolicy
	if restartPolicy == "" {
		restartPolicy = domain.RestartPolicyUnlessStopped
	}

	query := `
		INSERT INTO services (
			id, project_id, name, type, deploy_token, source_type, source_config, image, command, args,
			env_vars, ports, volumes, domains, replicas,
			cpu_limit, memory_limit, restart_policy, health_check, cron_jobs, labels,
			status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.q.ExecContext(ctx, query,
		s.ID, s.ProjectID, s.Name, string(s.Type), s.DeployToken, string(s.SourceType), string(sourceCfgJSON), s.Image, s.Command, string(argsJSON),
		string(envVarsJSON), string(portsJSON), string(volumesJSON), string(domainsJSON), replicas,
		s.Resources.CPULimit, s.Resources.MemoryLimit, restartPolicy, string(healthJSON), string(cronJobsJSON), string(labelsJSON),
		string(s.Status), s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrAlreadyExists
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *Repository) GetService(ctx context.Context, id string) (*domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, name, type, deploy_token, source_type, source_config, image, command, args,
		       env_vars, ports, volumes, domains, replicas,
		       cpu_limit, memory_limit, restart_policy, health_check, cron_jobs, labels,
		       status, created_at, updated_at
		FROM services WHERE id = ?
	`
	row := r.q.QueryRowContext(ctx, query, id)
	return r.scanService(row)
}

func (r *Repository) GetServiceByName(ctx context.Context, projectID, name string) (*domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, name, type, deploy_token, source_type, source_config, image, command, args,
		       env_vars, ports, volumes, domains, replicas,
		       cpu_limit, memory_limit, restart_policy, health_check, cron_jobs, labels,
		       status, created_at, updated_at
		FROM services WHERE project_id = ? AND name = ?
	`
	row := r.q.QueryRowContext(ctx, query, projectID, name)
	return r.scanService(row)
}

func (r *Repository) GetServiceByDeployToken(ctx context.Context, token string) (*domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if strings.TrimSpace(token) == "" {
		return nil, domain.ErrNotFound
	}

	query := `
		SELECT id, project_id, name, type, deploy_token, source_type, source_config, image, command, args,
		       env_vars, ports, volumes, domains, replicas,
		       cpu_limit, memory_limit, restart_policy, health_check, cron_jobs, labels,
		       status, created_at, updated_at
		FROM services WHERE deploy_token = ?
	`
	row := r.q.QueryRowContext(ctx, query, token)
	return r.scanService(row)
}

func (r *Repository) ListServicesByProject(ctx context.Context, projectID string) ([]*domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, name, type, deploy_token, source_type, source_config, image, command, args,
		       env_vars, ports, volumes, domains, replicas,
		       cpu_limit, memory_limit, restart_policy, health_check, cron_jobs, labels,
		       status, created_at, updated_at
		FROM services WHERE project_id = ? ORDER BY created_at ASC
	`
	rows, err := r.q.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*domain.Service
	for rows.Next() {
		s, err := r.scanServiceRow(rows)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, rows.Err()
}

func (r *Repository) ListAllServices(ctx context.Context) ([]*domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, name, type, deploy_token, source_type, source_config, image, command, args,
		       env_vars, ports, volumes, domains, replicas,
		       cpu_limit, memory_limit, restart_policy, health_check, cron_jobs, labels,
		       status, created_at, updated_at
		FROM services ORDER BY created_at ASC
	`
	rows, err := r.q.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*domain.Service
	for rows.Next() {
		s, err := r.scanServiceRow(rows)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, rows.Err()
}

func (r *Repository) UpdateService(ctx context.Context, s *domain.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := s.Validate(); err != nil {
		return err
	}

	s.UpdatedAt = time.Now().UTC()
	argsJSON, _ := json.Marshal(s.Args)
	envVarsJSON, _ := json.Marshal(s.EnvVars)
	portsJSON, _ := json.Marshal(s.Ports)
	volumesJSON, _ := json.Marshal(s.Volumes)
	domainsJSON, _ := json.Marshal(s.Domains)
	sourceCfgJSON, _ := json.Marshal(s.SourceConfig)
	healthJSON, _ := json.Marshal(s.HealthCheck)
	cronJobsJSON, _ := json.Marshal(s.CronJobs)
	labelsJSON, _ := json.Marshal(s.Labels)

	replicas := s.Replicas
	if replicas <= 0 {
		replicas = 1
	}
	restartPolicy := s.RestartPolicy
	if restartPolicy == "" {
		restartPolicy = domain.RestartPolicyUnlessStopped
	}

	query := `
		UPDATE services SET
			name = ?, type = ?, deploy_token = ?, source_type = ?, source_config = ?, image = ?, command = ?, args = ?,
			env_vars = ?, ports = ?, volumes = ?, domains = ?, replicas = ?,
			cpu_limit = ?, memory_limit = ?, restart_policy = ?, health_check = ?, cron_jobs = ?, labels = ?,
			status = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.q.ExecContext(ctx, query,
		s.Name, string(s.Type), s.DeployToken, string(s.SourceType), string(sourceCfgJSON), s.Image, s.Command, string(argsJSON),
		string(envVarsJSON), string(portsJSON), string(volumesJSON), string(domainsJSON), replicas,
		s.Resources.CPULimit, s.Resources.MemoryLimit, restartPolicy, string(healthJSON), string(cronJobsJSON), string(labelsJSON),
		string(s.Status), s.UpdatedAt, s.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrAlreadyExists
		}
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteService(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM services WHERE id = ?`
	res, err := r.q.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}
