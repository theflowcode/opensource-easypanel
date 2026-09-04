package http

import (
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type loginInput struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

// registerAuthRoutes binds authentication procedures to the oRPC dispatcher.
func registerAuthRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	d.Register("auth/login", func(c *orpc.Context) (any, error) {
		in, err := orpc.Bind[loginInput](c)
		if err != nil {
			return nil, err
		}

		email := strings.TrimSpace(in.Email)
		if email == "" || in.Password == "" {
			return nil, orpc.NewBadRequest("email and password are required")
		}

		user, err := db.GetUserByEmail(c.Context, email)
		if err != nil {
			return nil, orpc.NewBadRequest("invalid email or password")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
			return nil, orpc.NewBadRequest("invalid email or password")
		}

		duration := 24 * time.Hour
		if in.RememberMe {
			duration = 30 * 24 * time.Hour
		}

		rawToken := generateSecureToken()
		session := &domain.Session{
			ID:        domain.NewID(),
			UserID:    user.ID,
			TokenHash: orpc.HashToken(rawToken),
			ExpiresAt: time.Now().UTC().Add(duration),
			CreatedAt: time.Now().UTC(),
		}

		if err := db.CreateSession(c.Context, session); err != nil {
			return nil, err
		}

		clientIP := "127.0.0.1"
		if c.Request != nil {
			clientIP = c.Request.RemoteAddr
			if fwd := c.Request.Header.Get("X-Forwarded-For"); fwd != "" {
				clientIP = strings.Split(fwd, ",")[0]
			}
		}

		_ = db.CreateAction(c.Context, &domain.Action{
			ID:          domain.NewID(),
			Type:        domain.ActionTypeAuth,
			Status:      domain.ActionStatusDone,
			Description: "User " + user.Email + " logged in from " + clientIP,
			UserID:      user.ID,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		})

		return map[string]any{
			"token":            rawToken,
			"twoFactorEnabled": false,
		}, nil
	})

	d.Register("auth/getSession", func(c *orpc.Context) (any, error) {
		if c.Session == nil || c.User == nil || c.Session.IsExpired() {
			return map[string]any{}, nil
		}

		return map[string]any{
			"id":        c.Session.ID,
			"userId":    c.Session.UserID,
			"createdAt": c.Session.CreatedAt.Format(time.RFC3339),
			"expiresAt": c.Session.ExpiresAt.Format(time.RFC3339),
			"demoMode":  false,
		}, nil
	})

	d.Register("auth/getUser", func(c *orpc.Context) (any, error) {
		if c.User == nil || c.Session == nil || c.Session.IsExpired() {
			return nil, nil
		}

		return map[string]any{
			"id":        c.User.ID,
			"email":     c.User.Email,
			"admin":     c.User.Role == domain.RoleAdmin,
			"role":      c.User.Role,
			"createdAt": c.User.CreatedAt.Format(time.RFC3339),
		}, nil
	})

	d.Register("auth/logout", func(c *orpc.Context) (any, error) {
		if c.Session != nil {
			_ = db.DeleteSession(c.Context, c.Session.ID)
		}
		return map[string]bool{"success": true}, nil
	})
}
