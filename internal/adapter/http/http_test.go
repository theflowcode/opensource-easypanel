package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	openpanelhttp "github.com/opensource-easypanel/openpanel/internal/adapter/http"
	"github.com/opensource-easypanel/openpanel/internal/adapter/noop"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/test/mock"
)

func reqCtx() context.Context {
	return context.Background()
}

func setupTestServer() (*openpanelhttp.Server, *mock.MockDatabasePort) {
	mockDB := mock.NewMockDatabasePort()
	mockDocker := noop.NewNoOpDocker()
	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<!DOCTYPE html><html><body>OpenSource Easypanel</body></html>"))
	})
	server := openpanelhttp.NewServer(openpanelhttp.ServerDependencies{
		DB:         mockDB,
		Docker:     mockDocker,
		SPAHandler: spaHandler,
	})
	return server, mockDB
}

func callRPC(server http.Handler, path, token string, input any) *httptest.ResponseRecorder {
	var body bytes.Buffer
	if input != nil {
		_ = json.NewEncoder(&body).Encode(map[string]any{"json": input})
	} else {
		body.WriteString("{}")
	}
	req := httptest.NewRequest(http.MethodPost, "/api/rpc/"+path, &body)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	return w
}

func TestSetupAndAuthFlow(t *testing.T) {
	srv, _ := setupTestServer()

	// 1. Initial setup status: isComplete should be false
	w1 := callRPC(srv, "setup/getStatus", "", nil)
	if w1.Code != http.StatusOK {
		t.Fatalf("setup/getStatus failed: %d", w1.Code)
	}
	var res1 struct {
		JSON struct {
			IsComplete bool `json:"isComplete"`
		} `json:"json"`
	}
	_ = json.NewDecoder(w1.Body).Decode(&res1)
	if res1.JSON.IsComplete {
		t.Fatalf("expected isComplete to be false before setup")
	}

	// 2. Perform setup
	wSetup := callRPC(srv, "setup/setup", "", map[string]string{
		"email":    "admin@easypanel.io",
		"password": "supersecretpassword",
	})
	if wSetup.Code != http.StatusOK {
		t.Fatalf("setup/setup failed: %d: %s", wSetup.Code, wSetup.Body.String())
	}
	var setupRes struct {
		JSON struct {
			Token string `json:"token"`
		} `json:"json"`
	}
	_ = json.NewDecoder(wSetup.Body).Decode(&setupRes)
	token := setupRes.JSON.Token
	if token == "" {
		t.Fatalf("expected setup to return an auth token")
	}

	// 3. Post-setup status: isComplete should be true
	w2 := callRPC(srv, "setup/getStatus", "", nil)
	var res2 struct {
		JSON struct {
			IsComplete bool `json:"isComplete"`
		} `json:"json"`
	}
	_ = json.NewDecoder(w2.Body).Decode(&res2)
	if !res2.JSON.IsComplete {
		t.Fatalf("expected isComplete to be true after setup")
	}

	// 4. Duplicate setup must be rejected
	wDup := callRPC(srv, "setup/setup", "", map[string]string{
		"email":    "attacker@easypanel.io",
		"password": "password12345",
	})
	if wDup.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on duplicate setup, got %d", wDup.Code)
	}

	// 5. Test auth/getUser with valid token
	wUser := callRPC(srv, "auth/getUser", token, nil)
	if wUser.Code != http.StatusOK {
		t.Fatalf("auth/getUser failed: %d", wUser.Code)
	}
	var userRes struct {
		JSON *struct {
			Email string `json:"email"`
			Admin bool   `json:"admin"`
		} `json:"json"`
	}
	_ = json.NewDecoder(wUser.Body).Decode(&userRes)
	if userRes.JSON == nil || userRes.JSON.Email != "admin@easypanel.io" || !userRes.JSON.Admin {
		t.Fatalf("unexpected auth/getUser response: %+v", userRes.JSON)
	}

	// 6. Test auth/login
	wLogin := callRPC(srv, "auth/login", "", map[string]any{
		"email":      "admin@easypanel.io",
		"password":   "supersecretpassword",
		"rememberMe": true,
	})
	if wLogin.Code != http.StatusOK {
		t.Fatalf("auth/login failed: %d: %s", wLogin.Code, wLogin.Body.String())
	}

	// 7. Test auth/getSession
	wSess := callRPC(srv, "auth/getSession", token, nil)
	if wSess.Code != http.StatusOK {
		t.Fatalf("auth/getSession failed: %d", wSess.Code)
	}

	// 8. Test auth/logout
	wLogout := callRPC(srv, "auth/logout", token, nil)
	if wLogout.Code != http.StatusOK {
		t.Fatalf("auth/logout failed: %d", wLogout.Code)
	}
}

func TestLicensingAndBranding(t *testing.T) {
	srv, _ := setupTestServer()

	// 1. Verify Portal License unlocks Pro
	wPortal := callRPC(srv, "portalLicense/getLicensePayload", "", nil)
	if wPortal.Code != http.StatusOK {
		t.Fatalf("portalLicense failed: %d", wPortal.Code)
	}
	var portalRes struct {
		JSON struct {
			Valid bool `json:"valid"`
			Plan  struct {
				Name    string          `json:"name"`
				Options map[string]bool `json:"options"`
			} `json:"plan"`
		} `json:"json"`
	}
	_ = json.NewDecoder(wPortal.Body).Decode(&portalRes)
	if !portalRes.JSON.Valid || !portalRes.JSON.Plan.Options["advanced_monitoring"] {
		t.Fatalf("expected full pro features unlocked: %+v", portalRes.JSON)
	}

	// 2. Verify Lemon License
	wLemon := callRPC(srv, "lemonLicense/getLicensePayload", "", nil)
	if wLemon.Code != http.StatusOK {
		t.Fatalf("lemonLicense failed: %d", wLemon.Code)
	}

	// 3. Verify Branding
	wBrand := callRPC(srv, "branding/getBasicSettings", "", nil)
	if wBrand.Code != http.StatusOK {
		t.Fatalf("branding/getBasicSettings failed: %d", wBrand.Code)
	}
}

func TestProjectsAndServices(t *testing.T) {
	srv, mockDB := setupTestServer()

	// Perform setup to get admin session
	wSetup := callRPC(srv, "setup/setup", "", map[string]string{
		"email":    "admin@easypanel.io",
		"password": "supersecretpassword",
	})
	var setupRes struct {
		JSON struct {
			Token string `json:"token"`
		} `json:"json"`
	}
	_ = json.NewDecoder(wSetup.Body).Decode(&setupRes)
	token := setupRes.JSON.Token

	// 1. Check canCreateProject
	wCan := callRPC(srv, "projects/canCreateProject", token, nil)
	if wCan.Code != http.StatusOK {
		t.Fatalf("canCreateProject failed: %d", wCan.Code)
	}

	// 2. Create Project
	wCreate := callRPC(srv, "projects/createProject", token, map[string]string{
		"name":        "my-website",
		"description": "Production Website",
	})
	if wCreate.Code != http.StatusOK {
		t.Fatalf("projects/createProject failed: %d: %s", wCreate.Code, wCreate.Body.String())
	}

	// Create dummy service under project in DB
	proj, _ := mockDB.GetProjectByName(reqCtx(), "my-website")
	svc := &domain.Service{
		ID:          domain.NewID(),
		ProjectID:   proj.ID,
		ProjectName: proj.Name,
		Name:        "web",
		Type:        domain.ServiceTypeApp,
		Image:       "nginx:alpine",
		Replicas:    1,
		Status:      domain.ServiceStatusRunning,
	}
	_ = mockDB.CreateService(reqCtx(), svc)

	// 3. List Projects and Services
	wList := callRPC(srv, "projects/listProjectsAndServices", token, nil)
	if wList.Code != http.StatusOK {
		t.Fatalf("listProjectsAndServices failed: %d", wList.Code)
	}
	var listRes struct {
		JSON struct {
			Projects []struct {
				Name string `json:"name"`
			} `json:"projects"`
			Services []struct {
				ProjectName string `json:"projectName"`
				Name        string `json:"name"`
			} `json:"services"`
		} `json:"json"`
	}
	_ = json.NewDecoder(wList.Body).Decode(&listRes)
	if len(listRes.JSON.Projects) != 1 || len(listRes.JSON.Services) != 1 {
		t.Fatalf("expected 1 project and 1 service: %+v", listRes.JSON)
	}

	// 4. Inspect Service
	wInspect := callRPC(srv, "services/app/inspectService", token, map[string]string{
		"projectName": "my-website",
		"serviceName": "web",
	})
	if wInspect.Code != http.StatusOK {
		t.Fatalf("services/app/inspectService failed: %d: %s", wInspect.Code, wInspect.Body.String())
	}

	// 5. Update Service Env
	wEnv := callRPC(srv, "services/app/updateEnv", token, map[string]string{
		"projectName": "my-website",
		"serviceName": "web",
		"env":         "NODE_ENV=production\nPORT=3000",
	})
	if wEnv.Code != http.StatusOK {
		t.Fatalf("services/app/updateEnv failed: %d", wEnv.Code)
	}
}

func TestStaticSPAFallbackAndCORS(t *testing.T) {
	srv, _ := setupTestServer()

	// 1. SPA fallback for client-side route
	req := httptest.NewRequest(http.MethodGet, "/projects/my-website", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for SPA route, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("OpenSource Easypanel")) {
		t.Fatalf("expected SPA index HTML body")
	}

	// 2. CORS preflight
	reqOptions := httptest.NewRequest(http.MethodOptions, "/api/rpc/setup/getStatus", nil)
	reqOptions.Header.Set("Origin", "http://localhost:3000")
	wOptions := httptest.NewRecorder()
	srv.ServeHTTP(wOptions, reqOptions)

	if wOptions.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS, got %d", wOptions.Code)
	}
	if wOptions.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("unexpected CORS header: %s", wOptions.Header().Get("Access-Control-Allow-Origin"))
	}
}
