package http

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

// generateSecureToken creates a cryptographically random 32-byte hex token.
func generateSecureToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failure: " + err.Error())
	}
	return hex.EncodeToString(b)
}

type setupInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// registerSetupRoutes binds setup router procedures to the oRPC dispatcher.
func registerSetupRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	d.Register("setup/getStatus", func(c *orpc.Context) (any, error) {
		users, err := db.ListUsers(c.Context)
		if err != nil {
			return nil, err
		}
		hasAdmin := false
		for _, u := range users {
			if u.Role == domain.RoleAdmin {
				hasAdmin = true
				break
			}
		}
		return map[string]bool{"isComplete": hasAdmin}, nil
	})

	d.Register("setup/setup", func(c *orpc.Context) (any, error) {
		users, err := db.ListUsers(c.Context)
		if err != nil {
			return nil, err
		}
		if len(users) > 0 {
			return nil, orpc.NewBadRequest("server is already configured")
		}

		in, err := orpc.Bind[setupInput](c)
		if err != nil {
			return nil, err
		}

		email := strings.TrimSpace(in.Email)
		if email == "" || !strings.Contains(email, "@") {
			return nil, orpc.NewBadRequest("a valid email address is required")
		}
		if len(in.Password) < 6 {
			return nil, orpc.NewBadRequest("password must be at least 6 characters")
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, orpc.NewInternalError("failed to hash password: " + err.Error())
		}

		user := &domain.User{
			ID:           domain.NewID(),
			Email:        email,
			PasswordHash: string(hash),
			Role:         domain.RoleAdmin,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		if err := db.CreateUser(c.Context, user); err != nil {
			return nil, err
		}

		rawToken := generateSecureToken()
		session := &domain.Session{
			ID:        domain.NewID(),
			UserID:    user.ID,
			TokenHash: orpc.HashToken(rawToken),
			ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
			CreatedAt: time.Now().UTC(),
		}
		if err := db.CreateSession(c.Context, session); err != nil {
			return nil, err
		}

		return map[string]string{"token": rawToken}, nil
	})
}
