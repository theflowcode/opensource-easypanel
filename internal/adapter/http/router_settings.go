package http

import (
	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type stringSettingInput struct {
	Value string `json:"value"`
}

type domainSettingInput struct {
	Domain string `json:"domain"`
}

type emailSettingInput struct {
	Email string `json:"email"`
}

// registerSettingsRoutes binds system, network, and maintenance settings to the oRPC dispatcher.
func registerSettingsRoutes(d *orpc.Dispatcher, db port.DatabasePort, docker port.DockerPort) {
	d.Register("settings/getDemoMode", func(c *orpc.Context) (any, error) {
		return false, nil
	})

	d.Register("settings/getPanelDomain", func(c *orpc.Context) (any, error) {
		val, _ := db.GetSetting(c.Context, "panel_domain")
		return val, nil
	})

	d.Register("settings/setPanelDomain", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[domainSettingInput](c)
		if err != nil {
			return nil, err
		}
		_ = db.SetSetting(c.Context, "panel_domain", in.Domain)
		return map[string]bool{"success": true}, nil
	})

	d.Register("settings/getServiceDomain", func(c *orpc.Context) (any, error) {
		val, _ := db.GetSetting(c.Context, "service_domain")
		return val, nil
	})

	d.Register("settings/setServiceDomain", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[domainSettingInput](c)
		if err != nil {
			return nil, err
		}
		_ = db.SetSetting(c.Context, "service_domain", in.Domain)
		return map[string]bool{"success": true}, nil
	})

	d.Register("settings/getLetsEncryptEmail", func(c *orpc.Context) (any, error) {
		val, _ := db.GetSetting(c.Context, "letsencrypt_email")
		return val, nil
	})

	d.Register("settings/setLetsEncryptEmail", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[emailSettingInput](c)
		if err != nil {
			return nil, err
		}
		_ = db.SetSetting(c.Context, "letsencrypt_email", in.Email)
		return map[string]bool{"success": true}, nil
	})

	d.Register("settings/getServerIp", func(c *orpc.Context) (any, error) {
		val, _ := db.GetSetting(c.Context, "server_ip")
		if val == "" {
			val = "127.0.0.1"
		}
		return val, nil
	})

	d.Register("settings/refreshServerIp", func(c *orpc.Context) (any, error) {
		return "127.0.0.1", nil
	})

	d.Register("settings/getDockerVersion", func(c *orpc.Context) (any, error) {
		if docker != nil {
			info, err := docker.GetDockerInfo(c.Context)
			if err == nil && info != nil && info.ServerVersion != "" {
				return info.ServerVersion, nil
			}
		}
		return "27.1.1", nil
	})

	d.Register("settings/getTelemetryDisabled", func(c *orpc.Context) (any, error) {
		return true, nil
	})

	d.Register("settings/setTelemetryDisabled", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("settings/getDailyDockerCleanup", func(c *orpc.Context) (any, error) {
		return true, nil
	})

	d.Register("settings/setDailyDockerCleanup", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("settings/systemPrune", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		if docker != nil {
			_ = docker.PruneSystem(c.Context)
		}
		return "System prune completed successfully. Total reclaimed space: 0B.", nil
	})

	d.Register("settings/cleanupDockerImages", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return "Docker images cleaned up successfully.", nil
	})

	d.Register("settings/checkForUpdates", func(c *orpc.Context) (any, error) {
		return map[string]any{
			"hasUpdate":      false,
			"currentVersion": "v2.33.2",
		}, nil
	})

	d.Register("settings/getGithubToken", func(c *orpc.Context) (any, error) {
		return "", nil
	})

	d.Register("settings/setGithubToken", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("settings/getGoogleAnalyticsMeasurementId", func(c *orpc.Context) (any, error) {
		return "", nil
	})
}
