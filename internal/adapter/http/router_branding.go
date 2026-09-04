package http

import (
	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

// registerBrandingRoutes binds branding customization procedures to the oRPC dispatcher.
func registerBrandingRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	d.Register("branding/getInterfaceSettingsPublic", func(c *orpc.Context) (any, error) {
		lightLogo, _ := db.GetSetting(c.Context, "branding_light_logo")
		darkLogo, _ := db.GetSetting(c.Context, "branding_dark_logo")

		var lLogo, dLogo any
		if lightLogo != "" {
			lLogo = lightLogo
		}
		if darkLogo != "" {
			dLogo = darkLogo
		}

		return map[string]any{
			"lightLogo": lLogo,
			"darkLogo":  dLogo,
		}, nil
	})

	d.Register("branding/getOtherLinksSettings", func(c *orpc.Context) (any, error) {
		val, _ := db.GetSetting(c.Context, "branding_hide_other_links")
		return map[string]bool{
			"hideOtherLinks": val == "true",
		}, nil
	})

	d.Register("branding/getBasicSettings", func(c *orpc.Context) (any, error) {
		title, _ := db.GetSetting(c.Context, "branding_title")
		if title == "" {
			title = "OpenSource Easypanel"
		}
		hideIp, _ := db.GetSetting(c.Context, "branding_hide_ip")

		return map[string]any{
			"title":   title,
			"favicon": nil,
			"hideIp":  hideIp == "true",
		}, nil
	})

	d.Register("branding/getLogoSettings", func(c *orpc.Context) (any, error) {
		lightLogo, _ := db.GetSetting(c.Context, "branding_light_logo")
		darkLogo, _ := db.GetSetting(c.Context, "branding_dark_logo")
		return map[string]any{
			"lightLogo": lightLogo,
			"darkLogo":  darkLogo,
		}, nil
	})

	d.Register("branding/getLinksSettings", func(c *orpc.Context) (any, error) {
		return map[string]any{"links": []string{}}, nil
	})

	d.Register("branding/getCustomCodeSettings", func(c *orpc.Context) (any, error) {
		return map[string]string{"header": "", "body": ""}, nil
	})

	d.Register("branding/getErrorPageSettings", func(c *orpc.Context) (any, error) {
		return map[string]string{"errorPage": ""}, nil
	})

	// Setters (require Admin)
	d.Register("branding/setBasicSettings", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("branding/setLogoSettings", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("branding/setLinksSettings", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("branding/setCustomCodeSettings", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("branding/setErrorPageSettings", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})
}
