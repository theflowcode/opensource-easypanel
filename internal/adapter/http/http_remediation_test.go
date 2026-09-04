package http_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	openpanelhttp "github.com/opensource-easypanel/openpanel/internal/adapter/http"
	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/adapter/noop"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/test/mock"
)

func setupRemediationTestServer() (*openpanelhttp.Server, *mock.MockDatabasePort, *mock.MockProxyDriverPort) {
	mockDB := mock.NewMockDatabasePort()
	mockDocker := noop.NewNoOpDocker()
	mockProxy := mock.NewMockProxyDriverPort()
	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := openpanelhttp.NewServer(openpanelhttp.ServerDependencies{
		DB:         mockDB,
		Docker:     mockDocker,
		Proxy:      mockProxy,
		SPAHandler: spaHandler,
	})
	return server, mockDB, mockProxy
}

func TestRemediation_RBACEnforcement(t *testing.T) {
	srv, db, _ := setupRemediationTestServer()
	adminToken := loginAdmin(t, srv)

	// Create regular non-admin user (viewer)
	viewer := &domain.User{
		ID:           "user-viewer",
		Email:        "viewer@openpanel.dev",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456", // dummy
		Role:         domain.RoleViewer,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_ = db.CreateUser(reqCtx(), viewer)

	rawViewerToken := "viewer-token-secret"
	viewerSess := &domain.Session{
		ID:        "sess-viewer",
		UserID:    viewer.ID,
		TokenHash: orpc.HashToken(rawViewerToken),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	_ = db.CreateSession(reqCtx(), viewerSess)

	// Mutating route should fail with 401 unauthenticated
	wNoAuth := callRPC(srv, "users/createUser", "", map[string]any{
		"email":    "test@test.com",
		"password": "pass",
	})
	if wNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated user creation, got %d", wNoAuth.Code)
	}

	// Mutating route should fail with 403 forbidden for viewer role
	wViewer := callRPC(srv, "users/createUser", rawViewerToken, map[string]any{
		"email":    "test@test.com",
		"password": "pass",
	})
	if wViewer.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for viewer user creation, got %d", wViewer.Code)
	}

	// Admin succeeds
	wAdmin := callRPC(srv, "users/createUser", adminToken, map[string]any{
		"email":    "newadmin@openpanel.dev",
		"password": "AdminPassword123!",
		"role":     "admin",
	})
	if wAdmin.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin user creation, got %d: %s", wAdmin.Code, wAdmin.Body.String())
	}
}

func TestRemediation_DomainUpdateAndProxy(t *testing.T) {
	srv, db, mockProxy := setupRemediationTestServer()
	token := loginAdmin(t, srv)

	proj := &domain.Project{ID: "p1", Name: "myproj", CreatedAt: time.Now().UTC()}
	_ = db.CreateProject(reqCtx(), proj)
	svc := &domain.Service{ID: "s1", ProjectID: proj.ID, ProjectName: proj.Name, Name: "web", Type: domain.ServiceTypeApp, Image: "nginx:alpine", CreatedAt: time.Now().UTC()}
	_ = db.CreateService(reqCtx(), svc)

	// Create domain
	wCreate := callRPC(srv, "domains/createDomain", token, map[string]any{
		"projectName": "myproj",
		"serviceName": "web",
		"host":        "test.openpanel.dev",
		"port":        8080,
		"https":       true,
	})
	if wCreate.Code != http.StatusOK {
		t.Fatalf("createDomain failed: %d: %s", wCreate.Code, wCreate.Body.String())
	}

	var createRes struct {
		JSON struct {
			ID string `json:"id"`
		} `json:"json"`
	}
	_ = json.NewDecoder(wCreate.Body).Decode(&createRes)
	domID := createRes.JSON.ID
	if domID == "" {
		t.Fatalf("expected created domain id")
	}

	// Verify route applied to proxy
	if route, ok := mockProxy.Routes[svc.ID]; !ok || route.Domain != "test.openpanel.dev" {
		t.Fatalf("expected proxy route applied on createDomain, got: %+v", mockProxy.Routes)
	}

	// Update domain
	wUpdate := callRPC(srv, "domains/updateDomain", token, map[string]any{
		"id":          domID,
		"host":        "updated.openpanel.dev",
		"port":        9000,
		"path":        "/api",
		"https":       true,
		"middlewares": []string{"ratelimit-1"},
	})
	if wUpdate.Code != http.StatusOK {
		t.Fatalf("updateDomain failed: %d: %s", wUpdate.Code, wUpdate.Body.String())
	}

	// Check updated values in database
	updatedDom, err := db.GetDomain(reqCtx(), domID)
	if err != nil {
		t.Fatalf("failed to get domain from DB: %v", err)
	}
	if updatedDom.DomainName != "updated.openpanel.dev" || updatedDom.Port != 9000 || updatedDom.Path != "/api" {
		t.Fatalf("unexpected domain fields in DB: %+v", updatedDom)
	}

	// Delete domain
	wDel := callRPC(srv, "domains/deleteDomain", token, map[string]any{
		"id": domID,
	})
	if wDel.Code != http.StatusOK {
		t.Fatalf("deleteDomain failed: %d: %s", wDel.Code, wDel.Body.String())
	}
	if _, err := db.GetDomain(reqCtx(), domID); err != domain.ErrNotFound {
		t.Fatalf("expected domain to be deleted, got err: %v", err)
	}
}

func TestRemediation_UserUpdateAndApiTokens(t *testing.T) {
	srv, db, _ := setupRemediationTestServer()
	token := loginAdmin(t, srv)

	// Create user
	wCreate := callRPC(srv, "users/createUser", token, map[string]any{
		"email":    "alice@openpanel.dev",
		"password": "InitialPassword123!",
		"role":     "viewer",
	})
	if wCreate.Code != http.StatusOK {
		t.Fatalf("createUser failed: %d", wCreate.Code)
	}
	var createRes struct {
		JSON struct {
			ID string `json:"id"`
		} `json:"json"`
	}
	_ = json.NewDecoder(wCreate.Body).Decode(&createRes)
	aliceID := createRes.JSON.ID

	// Update user
	wUpdate := callRPC(srv, "users/updateUser", token, map[string]any{
		"id":    aliceID,
		"email": "alice-updated@openpanel.dev",
		"role":  "admin",
	})
	if wUpdate.Code != http.StatusOK {
		t.Fatalf("updateUser failed: %d: %s", wUpdate.Code, wUpdate.Body.String())
	}

	u, err := db.GetUserByID(reqCtx(), aliceID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if u.Email != "alice-updated@openpanel.dev" || u.Role != "admin" {
		t.Fatalf("unexpected user fields: %+v", u)
	}

	// Generate API token
	wTok := callRPC(srv, "users/generateApiToken", token, map[string]any{"name": "cli"})
	if wTok.Code != http.StatusOK {
		t.Fatalf("generateApiToken failed: %d", wTok.Code)
	}
	var tokRes struct {
		JSON struct {
			ID    string `json:"id"`
			Token string `json:"token"`
		} `json:"json"`
	}
	_ = json.NewDecoder(wTok.Body).Decode(&tokRes)
	if tokRes.JSON.Token == "" || tokRes.JSON.ID == "" {
		t.Fatalf("missing token or id in response")
	}

	// Verify token works for authenticated call
	wAuth := callRPC(srv, "users/listUsers", tokRes.JSON.Token, nil)
	if wAuth.Code != http.StatusOK {
		t.Fatalf("expected 200 with generated API token, got %d: %s", wAuth.Code, wAuth.Body.String())
	}

	// Revoke token
	wRevoke := callRPC(srv, "users/revokeApiToken", token, map[string]any{"id": tokRes.JSON.ID})
	if wRevoke.Code != http.StatusOK {
		t.Fatalf("revokeApiToken failed: %d", wRevoke.Code)
	}

	// Verify revoked token fails
	wAfterRevoke := callRPC(srv, "users/listUsers", tokRes.JSON.Token, nil)
	if wAfterRevoke.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after revoking token, got %d", wAfterRevoke.Code)
	}
}
