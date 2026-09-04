package orpc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"


	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/test/mock"
)

type echoInput struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

func TestORPCDispatcher_BasicRouting(t *testing.T) {
	mockDB := mock.NewMockDatabasePort()
	d := orpc.NewDispatcher(mockDB)

	d.Register("test/echo", func(c *orpc.Context) (any, error) {
		input, err := orpc.Bind[echoInput](c)
		if err != nil {
			return nil, err
		}
		return map[string]any{"echo": input.Message, "count": input.Count}, nil
	})

	body := `{"json": {"message": "hello orpc", "count": 42}}`
	req := httptest.NewRequest(http.MethodPost, "/api/rpc/test/echo", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	d.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		JSON struct {
			Echo  string `json:"echo"`
			Count int    `json:"count"`
		} `json:"json"`
	}
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response envelope: %v", err)
	}

	if res.JSON.Echo != "hello orpc" || res.JSON.Count != 42 {
		t.Fatalf("unexpected echo response: %+v", res.JSON)
	}
}

func TestORPCDispatcher_RouteNormalization(t *testing.T) {
	mockDB := mock.NewMockDatabasePort()
	d := orpc.NewDispatcher(mockDB)

	d.Register("setup.getStatus", func(c *orpc.Context) (any, error) {
		return map[string]bool{"isComplete": true}, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/api/rpc/setup/getStatus", bytes.NewBufferString("{}"))
	w := httptest.NewRecorder()
	d.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 via normalized path, got %d", w.Code)
	}
}

func TestORPCDispatcher_ErrorHandling(t *testing.T) {
	mockDB := mock.NewMockDatabasePort()
	d := orpc.NewDispatcher(mockDB)

	d.Register("test/badRequest", func(c *orpc.Context) (any, error) {
		return nil, orpc.NewBadRequest("invalid payload field")
	})
	d.Register("test/notFound", func(c *orpc.Context) (any, error) {
		return nil, domain.ErrNotFound
	})

	// Test 400 Bad Request
	req400 := httptest.NewRequest(http.MethodPost, "/api/rpc/test/badRequest", bytes.NewBufferString("{}"))
	w400 := httptest.NewRecorder()
	d.ServeHTTP(w400, req400)

	if w400.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w400.Code)
	}

	var errEnv orpc.ErrorEnvelope
	if err := json.NewDecoder(w400.Body).Decode(&errEnv); err != nil {
		t.Fatalf("failed to decode error envelope: %v", err)
	}
	if errEnv.JSON.Code != orpc.CodeBadRequest || errEnv.JSON.Status != 400 {
		t.Fatalf("unexpected error envelope structure: %+v", errEnv.JSON)
	}

	// Test 404 Not Found
	req404 := httptest.NewRequest(http.MethodPost, "/api/rpc/test/notFound", bytes.NewBufferString("{}"))
	w404 := httptest.NewRecorder()
	d.ServeHTTP(w404, req404)

	if w404.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w404.Code)
	}

	// Test Unknown Procedure 404
	reqUnknown := httptest.NewRequest(http.MethodPost, "/api/rpc/unknown/procedure", bytes.NewBufferString("{}"))
	wUnknown := httptest.NewRecorder()
	d.ServeHTTP(wUnknown, reqUnknown)

	if wUnknown.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown procedure, got %d", wUnknown.Code)
	}
}

func TestORPCDispatcher_AuthenticationAndRoles(t *testing.T) {
	mockDB := mock.NewMockDatabasePort()
	d := orpc.NewDispatcher(mockDB)

	user := &domain.User{
		ID:           "u1",
		Email:        "admin@easypanel.io",
		PasswordHash: "hashed",
		Role:         domain.RoleAdmin,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_ = mockDB.CreateUser(reqCtx(), user)

	rawToken := "secret-auth-token-12345"
	session := &domain.Session{
		ID:        "s1",
		UserID:    "u1",
		TokenHash: orpc.HashToken(rawToken),
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	_ = mockDB.CreateSession(reqCtx(), session)

	d.Register("auth/protected", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]string{"user": c.User.Email}, nil
	})

	// Unauthorized without header
	reqNoAuth := httptest.NewRequest(http.MethodPost, "/api/rpc/auth/protected", bytes.NewBufferString("{}"))
	wNoAuth := httptest.NewRecorder()
	d.ServeHTTP(wNoAuth, reqNoAuth)

	if wNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", wNoAuth.Code)
	}

	// Authorized with Bearer token
	reqAuth := httptest.NewRequest(http.MethodPost, "/api/rpc/auth/protected", bytes.NewBufferString("{}"))
	reqAuth.Header.Set("Authorization", "Bearer "+rawToken)
	wAuth := httptest.NewRecorder()
	d.ServeHTTP(wAuth, reqAuth)

	if wAuth.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid bearer token, got %d: %s", wAuth.Code, wAuth.Body.String())
	}
}

func reqCtx() context.Context {
	return context.Background()
}

