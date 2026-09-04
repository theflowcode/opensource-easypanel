package http_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func loginAdmin(t *testing.T, srv http.Handler) string {
	t.Helper()
	_ = callRPC(srv, "setup/setup", "", map[string]any{
		"email":    "admin@openpanel.dev",
		"password": "Password123!",
	})
	w := callRPC(srv, "auth/login", "", map[string]any{
		"email":    "admin@openpanel.dev",
		"password": "Password123!",
	})
	var res struct {
		JSON struct {
			Token string `json:"token"`
		} `json:"json"`
	}
	_ = json.NewDecoder(w.Body).Decode(&res)
	return res.JSON.Token
}

func TestDomainsUIParity(t *testing.T) {
	srv, db := setupTestServer()
	token := loginAdmin(t, srv)

	proj := &domain.Project{ID: "proj-1", Name: "myproject", CreatedAt: time.Now()}
	_ = db.CreateProject(reqCtx(), proj)
	svc := &domain.Service{ID: "svc-1", ProjectID: proj.ID, ProjectName: proj.Name, Name: "web", Type: domain.ServiceTypeApp, Image: "nginx:alpine", CreatedAt: time.Now()}
	_ = db.CreateService(reqCtx(), svc)

	// Create domain
	w1 := callRPC(srv, "domains/createDomain", token, map[string]any{
		"projectName": "myproject",
		"serviceName": "web",
		"domainName":  "app.example.com",
		"port":        8080,
	})
	if w1.Code != http.StatusOK {
		t.Fatalf("createDomain failed: %d, body: %s", w1.Code, w1.Body.String())
	}

	// Global listDomains
	w2 := callRPC(srv, "domains/listDomains", token, map[string]any{})
	if w2.Code != http.StatusOK {
		t.Fatalf("listDomains failed: %d", w2.Code)
	}
	var res2 struct {
		JSON []map[string]any `json:"json"`
	}
	_ = json.NewDecoder(w2.Body).Decode(&res2)
	if len(res2.JSON) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(res2.JSON))
	}
	d := res2.JSON[0]
	if d["host"] != "app.example.com" {
		t.Errorf("expected host 'app.example.com', got %v", d["host"])
	}
	if d["wildcard"] != false {
		t.Errorf("expected wildcard false, got %v", d["wildcard"])
	}
	dest, ok := d["serviceDestination"].(map[string]any)
	if !ok || dest["serviceName"] != "web" {
		t.Errorf("expected serviceDestination with serviceName 'web', got %v", d["serviceDestination"])
	}
}

func TestPortsAndMountsSubrouters(t *testing.T) {
	srv, db := setupTestServer()
	token := loginAdmin(t, srv)

	proj := &domain.Project{ID: "proj-pm", Name: "pmproj", CreatedAt: time.Now()}
	_ = db.CreateProject(reqCtx(), proj)
	svc := &domain.Service{ID: "svc-pm", ProjectID: proj.ID, ProjectName: proj.Name, Name: "backend", Type: domain.ServiceTypeApp, Image: "node:18-alpine", CreatedAt: time.Now()}
	_ = db.CreateService(reqCtx(), svc)

	// Create Port
	w1 := callRPC(srv, "ports/createPort", token, map[string]any{
		"projectName":   "pmproj",
		"serviceName":   "backend",
		"hostPort":      8080,
		"containerPort": 80,
		"protocol":      "tcp",
	})
	if w1.Code != http.StatusOK {
		t.Fatalf("createPort failed: %d", w1.Code)
	}

	// List Ports
	w2 := callRPC(srv, "ports/listPorts", token, map[string]any{
		"projectName": "pmproj",
		"serviceName": "backend",
	})
	if w2.Code != http.StatusOK {
		t.Fatalf("listPorts failed: %d", w2.Code)
	}
	var res2 struct {
		JSON []domain.PortMapping `json:"json"`
	}
	_ = json.NewDecoder(w2.Body).Decode(&res2)
	if len(res2.JSON) != 1 || res2.JSON[0].HostPort != 8080 {
		t.Fatalf("unexpected port list: %+v", res2.JSON)
	}

	// Create Mount
	w3 := callRPC(srv, "mounts/createMount", token, map[string]any{
		"projectName":   "pmproj",
		"serviceName":   "backend",
		"type":          "volume",
		"name":          "data-vol",
		"containerPath": "/var/data",
	})
	if w3.Code != http.StatusOK {
		t.Fatalf("createMount failed: %d", w3.Code)
	}

	// List Mounts
	w4 := callRPC(srv, "mounts/listMounts", token, map[string]any{
		"projectName": "pmproj",
		"serviceName": "backend",
	})
	if w4.Code != http.StatusOK {
		t.Fatalf("listMounts failed: %d", w4.Code)
	}
	var res4 struct {
		JSON []domain.VolumeMount `json:"json"`
	}
	_ = json.NewDecoder(w4.Body).Decode(&res4)
	if len(res4.JSON) != 1 || res4.JSON[0].Name != "data-vol" {
		t.Fatalf("unexpected mounts list: %+v", res4.JSON)
	}
}

func TestTelemetryAndMonitorSubrouters(t *testing.T) {
	srv, db := setupTestServer()
	token := loginAdmin(t, srv)

	proj := &domain.Project{ID: "proj-t", Name: "tproj", CreatedAt: time.Now()}
	_ = db.CreateProject(reqCtx(), proj)
	svc := &domain.Service{ID: "svc-t", ProjectID: proj.ID, ProjectName: proj.Name, Name: "worker", Type: domain.ServiceTypeApp, Image: "python:3.11-alpine", CreatedAt: time.Now()}
	_ = db.CreateService(reqCtx(), svc)

	// 1. metrics/getSystemStats
	w1 := callRPC(srv, "metrics/getSystemStats", token, nil)
	if w1.Code != http.StatusOK {
		t.Fatalf("metrics/getSystemStats failed: %d", w1.Code)
	}
	var res1 struct {
		JSON struct {
			CPU        [][2]any `json:"cpu"`
			Memory     [][2]any `json:"memory"`
			Disk       [][2]any `json:"disk"`
			NetworkIn  [][2]any `json:"networkIn"`
			NetworkOut [][2]any `json:"networkOut"`
			CPUCores   string   `json:"cpuCores"`
		} `json:"json"`
	}
	_ = json.NewDecoder(w1.Body).Decode(&res1)
	if len(res1.JSON.CPU) != 20 {
		t.Errorf("expected 20 cpu points, got %d", len(res1.JSON.CPU))
	}
	if res1.JSON.CPUCores == "" {
		t.Errorf("expected cpuCores to be set")
	}

	// 2. monitorOld/getMonitorTableData
	w2 := callRPC(srv, "monitorOld/getMonitorTableData", token, nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("monitorOld/getMonitorTableData failed: %d", w2.Code)
	}
	var res2 struct {
		JSON []map[string]any `json:"json"`
	}
	_ = json.NewDecoder(w2.Body).Decode(&res2)
	if len(res2.JSON) != 1 || res2.JSON[0]["serviceName"] != "worker" {
		t.Fatalf("unexpected monitor table rows: %+v", res2.JSON)
	}

	// 3. monitorOld/getDockerTaskStats
	w3 := callRPC(srv, "monitorOld/getDockerTaskStats", token, nil)
	if w3.Code != http.StatusOK {
		t.Fatalf("monitorOld/getDockerTaskStats failed: %d", w3.Code)
	}
}

func TestActionAuditLogging(t *testing.T) {
	srv, db := setupTestServer()
	token := loginAdmin(t, srv)

	proj := &domain.Project{ID: "proj-act", Name: "actproj", CreatedAt: time.Now()}
	_ = db.CreateProject(reqCtx(), proj)
	svc := &domain.Service{ID: "svc-act", ProjectID: proj.ID, ProjectName: proj.Name, Name: "frontend", Type: domain.ServiceTypeApp, Image: "nginx:alpine", CreatedAt: time.Now()}
	_ = db.CreateService(reqCtx(), svc)

	// Deploy service to trigger audit action
	w1 := callRPC(srv, "services/app/deployService", token, map[string]any{
		"projectName": "actproj",
		"serviceName": "frontend",
	})
	if w1.Code != http.StatusOK {
		t.Fatalf("deployService failed: %d", w1.Code)
	}

	// List actions
	w2 := callRPC(srv, "actions/listActions", token, map[string]any{
		"status": "done",
	})
	if w2.Code != http.StatusOK {
		t.Fatalf("listActions failed: %d", w2.Code)
	}
	var res2 struct {
		JSON []domain.Action `json:"json"`
	}
	_ = json.NewDecoder(w2.Body).Decode(&res2)
	// Must have at least auth and deployment actions
	if len(res2.JSON) < 2 {
		t.Fatalf("expected at least 2 actions, got %d", len(res2.JSON))
	}
}

func TestServerSettingsSubrouters(t *testing.T) {
	srv, _ := setupTestServer()
	token := loginAdmin(t, srv)

	// 1. Users CRUD
	w1 := callRPC(srv, "users/createUser", token, map[string]any{
		"email":    "viewer@example.com",
		"password": "ViewerPass123!",
		"role":     "viewer",
	})
	if w1.Code != http.StatusOK {
		t.Fatalf("createUser failed: %d", w1.Code)
	}

	w2 := callRPC(srv, "users/listUsers", token, nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("listUsers failed: %d", w2.Code)
	}
	var res2 struct {
		JSON []map[string]any `json:"json"`
	}
	_ = json.NewDecoder(w2.Body).Decode(&res2)
	if len(res2.JSON) != 2 {
		t.Fatalf("expected 2 users, got %d", len(res2.JSON))
	}

	// 2. Middlewares
	w3 := callRPC(srv, "middlewares/createMiddleware", token, map[string]any{
		"name": "basic-auth-mw",
		"type": "basicAuth",
		"config": map[string]any{
			"users": []string{"admin:$$apr1$$..."},
		},
	})
	if w3.Code != http.StatusOK {
		t.Fatalf("createMiddleware failed: %d", w3.Code)
	}

	w4 := callRPC(srv, "middlewares/listMiddlewares", token, nil)
	if w4.Code != http.StatusOK {
		t.Fatalf("listMiddlewares failed: %d", w4.Code)
	}

	// 3. Cluster and 2FA
	w5 := callRPC(srv, "cluster/listNodes", token, nil)
	if w5.Code != http.StatusOK {
		t.Fatalf("cluster/listNodes failed: %d", w5.Code)
	}

	w6 := callRPC(srv, "twoFactor/getStatus", token, nil)
	if w6.Code != http.StatusOK {
		t.Fatalf("twoFactor/getStatus failed: %d", w6.Code)
	}
}
