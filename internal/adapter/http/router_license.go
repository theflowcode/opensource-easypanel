package http

import (
	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
)

// registerLicenseRoutes binds license unlock procedures to the oRPC dispatcher.
// It bypasses all proprietary paywalls and unlocks 100% of Pro features for free.
func registerLicenseRoutes(d *orpc.Dispatcher) {
	d.Register("portalLicense/getLicensePayload", func(c *orpc.Context) (any, error) {
		return map[string]any{
			"valid": true,
			"type":  "portal",
			"plan": map[string]any{
				"name": "Pro",
				"options": map[string]bool{
					"advanced_monitoring": true,
					"notifications":       true,
					"branding":            true,
					"mutiple_users":       true,
					"access_control":      true,
					"cluster":             true,
				},
			},
		}, nil
	})

	d.Register("lemonLicense/getLicensePayload", func(c *orpc.Context) (any, error) {
		return map[string]any{
			"valid": true,
			"meta": map[string]any{
				"product_id": 1,
			},
		}, nil
	})

	d.Register("lemonLicense/getLicenseKey", func(c *orpc.Context) (any, error) {
		return "OPENSOURCE-EASYPANEL-PRO", nil
	})

	d.Register("portalLicense/activate", func(c *orpc.Context) (any, error) {
		return map[string]bool{"valid": true}, nil
	})

	d.Register("portalLicense/deactivate", func(c *orpc.Context) (any, error) {
		return map[string]bool{"valid": true}, nil
	})

	d.Register("lemonLicense/activate", func(c *orpc.Context) (any, error) {
		return map[string]bool{"valid": true}, nil
	})

	d.Register("lemonLicense/activateByOrder", func(c *orpc.Context) (any, error) {
		return map[string]bool{"valid": true}, nil
	})

	d.Register("lemonLicense/deactivate", func(c *orpc.Context) (any, error) {
		return map[string]bool{"valid": true}, nil
	})
}
