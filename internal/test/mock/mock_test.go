package mock_test

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
	"github.com/opensource-easypanel/openpanel/internal/test/mock"
)

// Compile-time interface adherence check
var (
	_ port.DockerPort      = (*mock.MockDockerPort)(nil)
	_ port.ProxyDriverPort = (*mock.MockProxyDriverPort)(nil)
	_ port.DatabasePort    = (*mock.MockDatabasePort)(nil)
	_ port.StreamPort      = (*mock.MockStreamPort)(nil)
	_ port.TemplatePort    = (*mock.MockTemplatePort)(nil)
	_ port.EventBusPort    = (*mock.MockEventBusPort)(nil)
)

func TestMockDockerPort(t *testing.T) {
	ctx := context.Background()
	m := mock.NewMockDockerPort()

	spec := domain.ServiceSpec{
		ID:        "srv-1",
		ProjectID: "p-1",
		Name:      "web",
		Type:      domain.ServiceTypeApp,
		Image:     "nginx:alpine",
		Replicas:  1,
	}

	res, err := m.DeployService(ctx, spec)
	if err != nil {
		t.Fatalf("DeployService failed: %v", err)
	}
	if res.ServiceID != "srv-1" {
		t.Errorf("DeployService returned unexpected result: %+v", res)
	}

	// Status
	st, err := m.GetServiceStatus(ctx, "srv-1")
	if err != nil || *st != domain.ServiceStatusRunning {
		t.Errorf("GetServiceStatus unexpected: status=%v, err=%v", st, err)
	}

	// Logs
	var buf bytes.Buffer
	if err := m.StreamServiceLogs(ctx, "srv-1", domain.LogStreamOptions{TailLines: 50}, &buf, &buf); err != nil {
		t.Fatalf("StreamServiceLogs failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Errorf("StreamServiceLogs buffer was empty")
	}

	// Build & Stats
	tag, err := m.BuildImage(ctx, domain.BuildConfig{ServiceID: "srv-1"}, nil)
	if err != nil || tag == "" {
		t.Errorf("BuildImage failed: %v", err)
	}
	if err := m.PullImage(ctx, "nginx:alpine", nil, nil); err != nil {
		t.Errorf("PullImage failed: %v", err)
	}
	stats, err := m.GetServiceStats(ctx, "srv-1")
	if err != nil || stats == nil {
		t.Errorf("GetServiceStats failed: %v", err)
	}
	if stats.NetworkInputBytes == 0 || stats.NetworkOutputBytes == 0 {
		t.Errorf("GetServiceStats missing network I/O telemetry: %+v", stats)
	}

	// Delete
	if err := m.DeleteService(ctx, "srv-1"); err != nil {
		t.Fatalf("DeleteService failed: %v", err)
	}
}

func TestMockProxyDriverPort(t *testing.T) {
	ctx := context.Background()
	m := mock.NewMockProxyDriverPort()

	route := domain.RouteConfig{
		ServiceID:   "srv-1",
		Domain:      "example.com",
		TargetPort:  80,
		EnableHTTPS: true,
	}

	if err := m.ApplyRoute(ctx, route); err != nil {
		t.Fatalf("ApplyRoute failed: %v", err)
	}
	if r, ok := m.Routes["srv-1"]; !ok || r.Domain != "example.com" {
		t.Errorf("Routes map unexpected: %+v", m.Routes)
	}

	if err := m.RemoveRoute(ctx, "srv-1"); err != nil {
		t.Fatalf("RemoveRoute failed: %v", err)
	}
	if _, ok := m.Routes["srv-1"]; ok {
		t.Errorf("expected route to be removed")
	}
}

func TestMockDatabasePort(t *testing.T) {
	ctx := context.Background()
	m := mock.NewMockDatabasePort()

	proj := &domain.Project{
		ID:   "p-1",
		Name: "Test Proj",
	}
	if err := m.CreateProject(ctx, proj); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	srv := &domain.Service{
		ID:        "s-1",
		ProjectID: "p-1",
		Name:      "web",
		Image:     "alpine:latest",
	}
	if err := m.CreateService(ctx, srv); err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	gotSrv, err := m.GetService(ctx, "s-1")
	if err != nil || gotSrv.Name != "web" {
		t.Errorf("GetService failed: got %+v, err=%v", gotSrv, err)
	}

	// Test cascade delete
	if err := m.DeleteProject(ctx, "p-1"); err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
	if _, err := m.GetService(ctx, "s-1"); err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound for cascade deleted service, got %v", err)
	}
}

func TestMockEventBusPort(t *testing.T) {
	ctx := context.Background()
	bus := mock.NewMockEventBusPort()

	var received []domain.Event
	var mu sync.Mutex

	_, err := bus.Subscribe(ctx, domain.EventServiceDeployed, func(e domain.Event) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, e)
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	ev := domain.Event{
		ID:        "evt-1",
		Type:      domain.EventServiceDeployed,
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"serviceId": "s-1"},
	}
	if err := bus.Publish(ctx, ev); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	mu.Lock()
	count := len(received)
	mu.Unlock()
	if count != 1 {
		t.Errorf("expected 1 received event, got %d", count)
	}
}

func TestMockTemplatePort(t *testing.T) {
	ctx := context.Background()
	tmplPort := mock.NewMockTemplatePort()

	templates, err := tmplPort.ListTemplates(ctx)
	if err != nil || len(templates) == 0 {
		t.Fatalf("ListTemplates failed: len=%d, err=%v", len(templates), err)
	}

	tmpl, err := tmplPort.GetTemplate(ctx, "postgres")
	if err != nil || tmpl.ID != "postgres" {
		t.Fatalf("GetTemplate failed: got %+v, err=%v", tmpl, err)
	}

	specs, err := tmplPort.RenderTemplate(ctx, tmpl, map[string]string{"POSTGRES_PASSWORD": "secretpassword"})
	if err != nil || len(specs) != 1 {
		t.Fatalf("RenderTemplate failed: len=%d, err=%v", len(specs), err)
	}
	if specs[0].EnvVars[0].Value != "secretpassword" {
		t.Errorf("RenderTemplate variable not interpolated: %s", specs[0].EnvVars[0].Value)
	}
}

func TestConcurrentMockAccess(t *testing.T) {
	ctx := context.Background()
	db := mock.NewMockDatabasePort()
	docker := mock.NewMockDockerPort()
	proxy := mock.NewMockProxyDriverPort()

	_ = db.CreateProject(ctx, &domain.Project{ID: "p-race", Name: "Race Project"})

	var wg sync.WaitGroup
	workers := 10

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id := fmt.Sprintf("s-%d", idx)
			s := &domain.Service{
				ID:        id,
				ProjectID: "p-race",
				Name:      fmt.Sprintf("web-%d", idx),
				Image:     "alpine",
			}
			_ = db.CreateService(ctx, s)
			_, _ = db.GetService(ctx, id)

			_, _ = docker.DeployService(ctx, s.ToSpec())
			_, _ = docker.GetServiceStatus(ctx, id)

			_ = proxy.ApplyRoute(ctx, domain.RouteConfig{
				ServiceID: id,
				Domain:    fmt.Sprintf("%s.example.com", id),
			})
			_ = proxy.RemoveRoute(ctx, id)
		}(i)
	}

	wg.Wait()
}

func TestMockDatabaseWithTx(t *testing.T) {
	ctx := context.Background()
	db := mock.NewMockDatabasePort()

	err := db.WithTx(ctx, func(tx port.DatabasePort) error {
		return tx.CreateProject(ctx, &domain.Project{ID: "p-tx", Name: "Tx Mock"})
	})
	if err != nil {
		t.Fatalf("WithTx failed: %v", err)
	}

	p, err := db.GetProject(ctx, "p-tx")
	if err != nil || p.Name != "Tx Mock" {
		t.Errorf("GetProject failed after WithTx: %+v, err=%v", p, err)
	}
}

func TestMockResetMethods(t *testing.T) {
	ctx := context.Background()

	db := mock.NewMockDatabasePort()
	_ = db.CreateProject(ctx, &domain.Project{ID: "p-1", Name: "P1"})
	db.Reset()
	if len(db.Projects) != 0 || len(db.Calls) != 0 {
		t.Error("MockDatabasePort.Reset() failed to clear state")
	}

	docker := mock.NewMockDockerPort()
	_, _ = docker.DeployService(ctx, domain.ServiceSpec{ID: "s-1"})
	docker.Reset()
	if len(docker.DeployedServices) != 0 || len(docker.Calls) != 0 {
		t.Error("MockDockerPort.Reset() failed to clear state")
	}

	proxy := mock.NewMockProxyDriverPort()
	_ = proxy.ApplyRoute(ctx, domain.RouteConfig{ServiceID: "s-1"})
	proxy.Reset()
	if len(proxy.Routes) != 0 || len(proxy.Calls) != 0 {
		t.Error("MockProxyDriverPort.Reset() failed to clear state")
	}

	stream := mock.NewMockStreamPort()
	var buf bytes.Buffer
	_ = stream.SubscribeLogs(ctx, "s-1", domain.LogStreamOptions{}, &buf)
	stream.Reset()
	if len(stream.Calls) != 0 {
		t.Error("MockStreamPort.Reset() failed to clear calls")
	}

	tmpl := mock.NewMockTemplatePort()
	_, _ = tmpl.ParseTemplate(ctx, []byte(`{}`))
	tmpl.Reset()
	if len(tmpl.Templates) != 0 || len(tmpl.Calls) != 0 {
		t.Error("MockTemplatePort.Reset() failed to clear state")
	}

	bus := mock.NewMockEventBusPort()
	_ = bus.Publish(ctx, domain.Event{ID: "e-1"})
	bus.Reset()
	if len(bus.Events) != 0 || len(bus.Calls) != 0 {
		t.Error("MockEventBusPort.Reset() failed to clear state")
	}
}
