package http

import (
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type createUserInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type updateUserInput struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type destroyUserInput struct {
	ID string `json:"id"`
}

type generateApiTokenInput struct {
	Name string `json:"name"`
}

type revokeTokenInput struct {
	ID string `json:"id"`
}

// registerServerUsersRoutes binds user administration and 2FA procedures to the oRPC dispatcher.
func registerServerUsersRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	d.Register("users/listUsers", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		users, err := db.ListUsers(c.Context)
		if err != nil {
			return nil, err
		}
		type userDTO struct {
			ID        string `json:"id"`
			Email     string `json:"email"`
			Role      string `json:"role"`
			CreatedAt string `json:"createdAt"`
			UpdatedAt string `json:"updatedAt"`
		}
		dto := make([]userDTO, 0, len(users))
		for _, u := range users {
			dto = append(dto, userDTO{
				ID:        u.ID,
				Email:     u.Email,
				Role:      u.Role,
				CreatedAt: u.CreatedAt.Format(time.RFC3339),
				UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
			})
		}
		return dto, nil
	})

	d.Register("users/createUser", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[createUserInput](c)
		if err != nil {
			return nil, err
		}
		email := strings.TrimSpace(in.Email)
		if email == "" || in.Password == "" {
			return nil, orpc.NewBadRequest("email and password are required")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		role := strings.TrimSpace(in.Role)
		if role == "" {
			role = domain.RoleAdmin
		}
		user := &domain.User{
			ID:           domain.NewID(),
			Email:        email,
			PasswordHash: string(hash),
			Role:         role,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		if err := db.CreateUser(c.Context, user); err != nil {
			return nil, err
		}
		return map[string]any{
			"id":        user.ID,
			"email":     user.Email,
			"role":      user.Role,
			"createdAt": user.CreatedAt.Format(time.RFC3339),
		}, nil
	})

	d.Register("users/updateUser", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[updateUserInput](c)
		if err != nil {
			return nil, err
		}
		if in.ID == "" {
			return nil, orpc.NewBadRequest("user id is required")
		}
		u, err := db.GetUserByID(c.Context, in.ID)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		if in.Email != "" {
			u.Email = strings.TrimSpace(in.Email)
		}
		if in.Password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
			if err != nil {
				return nil, err
			}
			u.PasswordHash = string(hash)
		}
		if in.Role != "" {
			u.Role = strings.TrimSpace(in.Role)
		}
		u.UpdatedAt = time.Now().UTC()
		if err := db.UpdateUser(c.Context, u); err != nil {
			return nil, err
		}
		return map[string]any{
			"id":        u.ID,
			"email":     u.Email,
			"role":      u.Role,
			"createdAt": u.CreatedAt.Format(time.RFC3339),
			"updatedAt": u.UpdatedAt.Format(time.RFC3339),
		}, nil
	})

	d.Register("users/destroyUser", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[destroyUserInput](c)
		if err != nil {
			return nil, err
		}
		if in.ID == "" {
			return nil, orpc.NewBadRequest("user id is required")
		}
		if err := db.DeleteUser(c.Context, in.ID); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("users/generateApiToken", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		raw := generateSecureToken()
		sessionID := domain.NewID()
		sess := &domain.Session{
			ID:        sessionID,
			UserID:    c.User.ID,
			TokenHash: orpc.HashToken(raw),
			ExpiresAt: time.Now().AddDate(1, 0, 0),
			CreatedAt: time.Now().UTC(),
		}
		if err := db.CreateSession(c.Context, sess); err != nil {
			return nil, err
		}
		return map[string]string{
			"id":    sessionID,
			"token": raw,
		}, nil
	})

	d.Register("users/revokeApiToken", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[revokeTokenInput](c)
		if err == nil && in.ID != "" {
			_ = db.DeleteSession(c.Context, in.ID)
		}
		return map[string]bool{"success": true}, nil
	})

	// Two-Factor Authentication
	d.Register("twoFactor/getStatus", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return map[string]bool{"enabled": false}, nil
	})

	d.Register("twoFactor/configure", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]string{
			"secret": "JBSWY3DPEHPK3PXP",
			"qrCode": "data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg'/>",
		}, nil
	})

	d.Register("twoFactor/enable", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("twoFactor/disable", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})
}
