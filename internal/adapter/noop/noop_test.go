package noop_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/opensource-easypanel/openpanel/internal/adapter/noop"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

// Compile-time checks
var (
	_ port.DockerPort      = (*noop.NoOpDocker)(nil)
	_ port.ProxyDriverPort = (*noop.NoOpProxyDriver)(nil)
	_ port.StreamPort      = (*noop.NoOpStreamer)(nil)
	_ port.TemplatePort    = (*noop.NoOpTemplate)(nil)
	_ port.EventBusPort    = (*noop.NoOpEventBus)(nil)
)

func TestNoOpAdapters(t *testing.T) {
	ctx := context.Background()

	// Docker
	d := noop.NewNoOpDocker()
	res, err := d.DeployService(ctx, domain.ServiceSpec{ID: "srv-1"})
	if err != nil || res.Status != "running" {
		t.Errorf("NoOpDocker.DeployService failed: %v", err)
	}
	if err := d.StopService(ctx, "srv-1"); err != nil {
		t.Errorf("NoOpDocker.StopService failed: %v", err)
	}
	if err := d.RestartService(ctx, "srv-1"); err != nil {
		t.Errorf("NoOpDocker.RestartService failed: %v", err)
	}
	if err := d.DeleteService(ctx, "srv-1"); err != nil {
		t.Errorf("NoOpDocker.DeleteService failed: %v", err)
	}
	tag, err := d.BuildImage(ctx, domain.BuildConfig{ServiceID: "srv-1"}, nil)
	if err != nil || tag == "" {
		t.Errorf("NoOpDocker.BuildImage failed: %v", err)
	}
	if err := d.PullImage(ctx, "alpine:latest", nil, nil); err != nil {
		t.Errorf("NoOpDocker.PullImage failed: %v", err)
	}
	st, err := d.GetServiceStatus(ctx, "srv-1")
	if err != nil || *st != domain.ServiceStatusRunning {
		t.Errorf("NoOpDocker.GetServiceStatus failed: %v", err)
	}
	stats, err := d.GetServiceStats(ctx, "srv-1")
	if err != nil || stats == nil {
		t.Errorf("NoOpDocker.GetServiceStats failed: %v", err)
	}
	info, err := d.GetDockerInfo(ctx)
	if err != nil || info == nil || info.ServerVersion == "" {
		t.Errorf("NoOpDocker.GetDockerInfo failed: %v", err)
	}
	metrics, err := d.GetHostMetrics(ctx)
	if err != nil || metrics == nil || metrics.MemoryTotalBytes == 0 {
		t.Errorf("NoOpDocker.GetHostMetrics failed: %v", err)
	}
	var buf bytes.Buffer
	if err := d.StreamServiceLogs(ctx, "srv-1", domain.LogStreamOptions{TailLines: 100}, &buf, &buf); err != nil {
		t.Errorf("NoOpDocker.StreamServiceLogs failed: %v", err)
	}
	if err := d.EnsureNetwork(ctx, "net"); err != nil {
		t.Errorf("NoOpDocker.EnsureNetwork failed: %v", err)
	}
	if err := d.EnsureVolume(ctx, "vol"); err != nil {
		t.Errorf("NoOpDocker.EnsureVolume failed: %v", err)
	}
	containers, err := d.ListContainers(ctx)
	if err != nil || len(containers) != 0 {
		t.Errorf("NoOpDocker.ListContainers failed: %v", err)
	}
	if err := d.PruneSystem(ctx); err != nil {
		t.Errorf("NoOpDocker.PruneSystem failed: %v", err)
	}

	// Proxy
	p := noop.NewNoOpProxyDriver()
	if err := p.ApplyRoute(ctx, domain.RouteConfig{ServiceID: "s-1"}); err != nil {
		t.Errorf("NoOpProxyDriver.ApplyRoute failed: %v", err)
	}
	if err := p.RemoveRoute(ctx, "s-1"); err != nil {
		t.Errorf("NoOpProxyDriver.RemoveRoute failed: %v", err)
	}
	if err := p.SyncAllRoutes(ctx, nil); err != nil {
		t.Errorf("NoOpProxyDriver.SyncAllRoutes failed: %v", err)
	}

	// Stream
	s := noop.NewNoOpStreamer()
	if err := s.SubscribeLogs(ctx, "s-1", domain.LogStreamOptions{Follow: true}, &buf); err != nil {
		t.Errorf("NoOpStreamer.SubscribeLogs failed: %v", err)
	}
	if err := s.HandleTerminalStream(ctx, "s-1", &buf, &buf, nil); err != nil {
		t.Errorf("NoOpStreamer.HandleTerminalStream failed: %v", err)
	}

	// Template
	tmpl := noop.NewNoOpTemplate()
	templates, err := tmpl.ListTemplates(ctx)
	if err != nil || len(templates) != 0 {
		t.Errorf("NoOpTemplate.ListTemplates failed: %v", err)
	}
	specs, err := tmpl.RenderTemplate(ctx, nil, nil)
	if err != nil || len(specs) != 0 {
		t.Errorf("NoOpTemplate.RenderTemplate failed: %v", err)
	}

	// EventBus
	bus := noop.NewNoOpEventBus()
	if err := bus.Publish(ctx, domain.Event{ID: "e-1"}); err != nil {
		t.Errorf("NoOpEventBus.Publish failed: %v", err)
	}
	sub, err := bus.Subscribe(ctx, "*", func(e domain.Event) {})
	if err != nil || sub == nil {
		t.Errorf("NoOpEventBus.Subscribe failed: %v", err)
	}
	sub.Unsubscribe()
}
