package orpc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

// HandlerFunc defines the signature for an oRPC procedure handler.
type HandlerFunc func(c *Context) (any, error)

// Dispatcher coordinates oRPC routing, request envelope decoding, and execution.
type Dispatcher struct {
	mu     sync.RWMutex
	routes map[string]HandlerFunc
	db     port.DatabasePort
}

// NewDispatcher creates an initialized oRPC Dispatcher.
func NewDispatcher(db port.DatabasePort) *Dispatcher {
	return &Dispatcher{
		routes: make(map[string]HandlerFunc),
		db:     db,
	}
}

// HashToken computes the SHA-256 hex digest of an authentication token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// NormalizePath standardizes procedure path formatting (e.g. "auth.login" -> "auth/login").
func NormalizePath(path string) string {
	p := strings.Trim(path, "/")
	p = strings.ReplaceAll(p, ".", "/")
	return p
}

// Register attaches a HandlerFunc to the given procedure path.
func (d *Dispatcher) Register(path string, handler HandlerFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.routes[NormalizePath(path)] = handler
}

// ServeHTTP handles incoming HTTP requests matching /api/rpc/*.
func (d *Dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/rpc")
	procPath := NormalizePath(path)

	d.mu.RLock()
	handler, exists := d.routes[procPath]
	d.mu.RUnlock()

	if !exists {
		WriteError(w, http.StatusNotFound, CodeNotFound, fmt.Sprintf("procedure %q not found", procPath))
		return
	}

	rawInput, err := d.extractInput(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, CodeBadRequest, "failed to parse request envelope: "+err.Error())
		return
	}

	token := d.extractToken(r)
	var user *domain.User
	var session *domain.Session

	if token != "" && d.db != nil {
		sess, sErr := d.db.GetSession(r.Context(), HashToken(token))
		if sErr == nil && sess != nil && !sess.IsExpired() {
			session = sess
			u, uErr := d.db.GetUserByID(r.Context(), sess.UserID)
			if uErr == nil {
				user = u
			}
		}
	}

	c := &Context{
		Context:        r.Context(),
		Request:        r,
		ResponseWriter: w,
		Token:          token,
		User:           user,
		Session:        session,
		Input:          rawInput,
	}

	res, err := handler(c)
	if err != nil {
		d.handleError(w, err)
		return
	}

	WriteSuccess(w, res)
}

func (d *Dispatcher) extractInput(r *http.Request) (json.RawMessage, error) {
	if r.Method == http.MethodGet {
		qData := r.URL.Query().Get("data")
		if qData == "" {
			return nil, nil
		}
		var env RequestEnvelope
		if err := json.Unmarshal([]byte(qData), &env); err == nil && len(env.JSON) > 0 {
			return env.JSON, nil
		}
		return json.RawMessage(qData), nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, nil
	}

	var env RequestEnvelope
	if err := json.Unmarshal(body, &env); err == nil && len(env.JSON) > 0 {
		return env.JSON, nil
	}
	return json.RawMessage(body), nil
}

func (d *Dispatcher) extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			return strings.TrimSpace(auth[7:])
		}
		return strings.TrimSpace(auth)
	}
	if ep := r.Header.Get("easypanel"); ep != "" {
		return strings.TrimSpace(ep)
	}
	return strings.TrimSpace(r.URL.Query().Get("token"))
}

func (d *Dispatcher) handleError(w http.ResponseWriter, err error) {
	var oe *ORPCError
	if errors.As(err, &oe) {
		WriteError(w, oe.Status, oe.Code, oe.Message)
		return
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		WriteError(w, http.StatusNotFound, CodeNotFound, err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, CodeUnauthorized, err.Error())
	case errors.Is(err, domain.ErrValidation):
		WriteError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		WriteError(w, http.StatusConflict, CodeConflict, err.Error())
	default:
		WriteError(w, http.StatusInternalServerError, CodeInternalServerError, err.Error())
	}
}
