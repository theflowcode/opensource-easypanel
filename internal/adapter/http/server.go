package http

import (
	"crypto/sha1"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/adapter/noop"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

// ServerDependencies aggregates all inbound & outbound ports for dependency injection.
type ServerDependencies struct {
	DB         port.DatabasePort
	Docker     port.DockerPort
	Proxy      port.ProxyDriverPort
	Stream     port.StreamPort
	Template   port.TemplatePort
	Notifier   port.NotifierPort
	EventBus   port.EventBusPort
	SPAHandler http.Handler
}

// Server coordinates the oRPC dispatcher, WebSocket endpoints, and embedded SPA routing.
type Server struct {
	deps       ServerDependencies
	dispatcher *orpc.Dispatcher
	mux        *http.ServeMux
}

// NewServer builds and wires the entire HTTP control plane API and frontend router.
func NewServer(deps ServerDependencies) *Server {
	if deps.Docker == nil {
		deps.Docker = noop.NewNoOpDocker()
	}
	if deps.Proxy == nil {
		deps.Proxy = noop.NewNoOpProxyDriver()
	}
	if deps.Stream == nil {
		deps.Stream = noop.NewNoOpStreamer()
	}
	if deps.Template == nil {
		deps.Template = noop.NewNoOpTemplate()
	}
	if deps.Notifier == nil {
		deps.Notifier = noop.NewNoOpNotifier()
	}
	if deps.EventBus == nil {
		deps.EventBus = noop.NewNoOpEventBus()
	}

	d := orpc.NewDispatcher(deps.DB)

	// Register all oRPC routers
	registerSetupRoutes(d, deps.DB)
	registerAuthRoutes(d, deps.DB)
	registerBrandingRoutes(d, deps.DB)
	registerLicenseRoutes(d)
	registerProjectsRoutes(d, deps.DB)
	registerServicesAppRoutes(d, deps.DB, deps.Docker)
	registerServicesAppConfigRoutes(d, deps.DB)
	registerServicesDBRoutes(d, deps.DB, deps.Docker)
	registerServicesCommonRoutes(d, deps.DB)
	registerPortsRoutes(d, deps.DB)
	registerMountsRoutes(d, deps.DB)
	registerDomainsRoutes(d, deps.DB, deps.Proxy)
	registerSettingsRoutes(d, deps.DB, deps.Docker)
	registerTelemetryRoutes(d, deps.DB, deps.Docker)
	registerServerUsersRoutes(d, deps.DB)
	registerServerInfraRoutes(d, deps.DB)
	registerActionsAndExtraRoutes(d, deps.DB)

	mux := http.NewServeMux()

	// Mount oRPC API dispatcher
	mux.Handle("/api/rpc/", d)

	// Mount WebSocket stubs
	wsEndpoints := []string{
		"/ws/dockerEvents",
		"/ws/diskUsage",
		"/ws/hostShell",
		"/ws/serviceLogs",
		"/ws/containerShell",
	}
	for _, ep := range wsEndpoints {
		mux.HandleFunc(ep, handleWebSocketStub)
	}

	// Mount SPA static files & client-side routing fallback
	if deps.SPAHandler != nil {
		mux.Handle("/", deps.SPAHandler)
	}

	return &Server{
		deps:       deps,
		dispatcher: d,
		mux:        mux,
	}
}

// Dispatcher returns the underlying oRPC dispatcher for direct procedure testing.
func (s *Server) Dispatcher() *orpc.Dispatcher {
	return s.dispatcher
}

// DB returns the database port.
func (s *Server) DB() port.DatabasePort {
	return s.deps.DB
}

// Docker returns the docker port.
func (s *Server) Docker() port.DockerPort {
	return s.deps.Docker
}

// Proxy returns the proxy driver port.
func (s *Server) Proxy() port.ProxyDriverPort {
	return s.deps.Proxy
}

func isAllowedOrigin(origin, reqHost string) bool {
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	oHost := u.Hostname()
	if oHost == "localhost" || oHost == "127.0.0.1" || oHost == "::1" {
		return true
	}
	reqHostname := reqHost
	if h, _, err := net.SplitHostPort(reqHost); err == nil {
		reqHostname = h
	}
	return strings.EqualFold(oHost, reqHostname)
}

// ServeHTTP applies secure CORS middleware and forwards requests to the multiplexer.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" && isAllowedOrigin(origin, r.Host) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, easypanel")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func handleWebSocketStub(w http.ResponseWriter, r *http.Request) {
	if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
		key := r.Header.Get("Sec-WebSocket-Key")
		if key == "" {
			http.Error(w, "missing Sec-WebSocket-Key", http.StatusBadRequest)
			return
		}
		h := sha1.New()
		h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
		acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "websocket hijack not supported", http.StatusInternalServerError)
			return
		}
		conn, bufrw, err := hj.Hijack()
		if err != nil {
			return
		}
		defer conn.Close()

		_, _ = bufrw.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
		_, _ = bufrw.WriteString("Upgrade: websocket\r\n")
		_, _ = bufrw.WriteString("Connection: Upgrade\r\n")
		_, _ = bufrw.WriteString("Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n")
		_ = bufrw.Flush()

		// Hold connection until client disconnects
		buf := make([]byte, 512)
		for {
			_, err := conn.Read(buf)
			if err != nil {
				break
			}
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ready"}`))
}
