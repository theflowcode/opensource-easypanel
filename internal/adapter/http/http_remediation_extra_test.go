package http_test

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemediation_StorageProvidersAndNotificationsParity(t *testing.T) {
	srv, _, _ := setupRemediationTestServer()
	token := loginAdmin(t, srv)

	// Test storage providers listOptions
	wOpts := callRPC(srv, "storageProviders/common/listOptions", token, nil)
	if wOpts.Code != http.StatusOK {
		t.Fatalf("listOptions failed: %d", wOpts.Code)
	}

	// Test provider-specific create (s3) and secret masking
	wCreateS3 := callRPC(srv, "storageProviders/s3/createProvider", token, map[string]any{
		"name":      "my-s3",
		"endpoint":  "https://s3.amazonaws.com",
		"bucket":    "backups",
		"accessKey": "AKIA12345",
		"secretKey": "TopSecretKey987!",
	})
	if wCreateS3.Code != http.StatusOK {
		t.Fatalf("createProvider (s3) failed: %d: %s", wCreateS3.Code, wCreateS3.Body.String())
	}
	var s3Res struct {
		JSON struct {
			ID        string `json:"id"`
			SecretKey string `json:"secretKey"`
		} `json:"json"`
	}
	_ = json.NewDecoder(wCreateS3.Body).Decode(&s3Res)
	if s3Res.JSON.SecretKey != "********" {
		t.Fatalf("expected masked secretKey, got: %q", s3Res.JSON.SecretKey)
	}

	// Test storageProviders/common/list alias
	wList := callRPC(srv, "storageProviders/common/list", token, nil)
	if wList.Code != http.StatusOK {
		t.Fatalf("storageProviders/common/list alias failed: %d", wList.Code)
	}

	// Test notifications/createNotificationChannel and masking
	wCreateNotif := callRPC(srv, "notifications/createNotificationChannel", token, map[string]any{
		"name": "Discord Alerts",
		"target": map[string]any{
			"type":     "discord",
			"url":      "https://discord.com/api/webhooks/123/abc",
			"password": "SuperSecretPassword",
		},
		"events": map[string]any{
			"appDeploy": map[string]any{"enabled": true},
		},
	})
	if wCreateNotif.Code != http.StatusOK {
		t.Fatalf("createNotificationChannel failed: %d: %s", wCreateNotif.Code, wCreateNotif.Body.String())
	}
	var notifRes struct {
		JSON struct {
			ID     string `json:"id"`
			Target struct {
				Password string `json:"password"`
			} `json:"target"`
		} `json:"json"`
	}
	_ = json.NewDecoder(wCreateNotif.Body).Decode(&notifRes)
	if notifRes.JSON.Target.Password != "********" {
		t.Fatalf("expected masked password in notification target, got %q", notifRes.JSON.Target.Password)
	}

	// Test notifications/sendTestNotification alias
	wTestNotif := callRPC(srv, "notifications/sendTestNotification", token, map[string]any{
		"id": notifRes.JSON.ID,
	})
	if wTestNotif.Code != http.StatusOK {
		t.Fatalf("sendTestNotification alias failed: %d", wTestNotif.Code)
	}
}

func TestRemediation_CORSOriginSecurity(t *testing.T) {
	srv, _, _ := setupRemediationTestServer()

	// 1. Untrusted origin should NOT be reflected with credentials
	reqUntrusted := httptest.NewRequest(http.MethodOptions, "/api/rpc/setup/getStatus", nil)
	reqUntrusted.Header.Set("Origin", "https://malicious-attacker.com")
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, reqUntrusted)

	if w1.Header().Get("Access-Control-Allow-Origin") == "https://malicious-attacker.com" {
		t.Fatalf("untrusted origin must not be reflected in Access-Control-Allow-Origin")
	}
	if w1.Header().Get("Access-Control-Allow-Credentials") == "true" {
		t.Fatalf("untrusted origin must not have Access-Control-Allow-Credentials: true")
	}

	// 2. Trusted localhost origin should be allowed with credentials
	reqTrusted := httptest.NewRequest(http.MethodOptions, "/api/rpc/setup/getStatus", nil)
	reqTrusted.Header.Set("Origin", "http://localhost:3000")
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, reqTrusted)

	if w2.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("trusted origin should be reflected, got: %s", w2.Header().Get("Access-Control-Allow-Origin"))
	}
	if w2.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("trusted origin should allow credentials")
	}
}

func TestRemediation_WebSocketHandshakeRFC6455(t *testing.T) {
	srv, _, _ := setupRemediationTestServer()
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// Non-websocket request returns 200 {"status":"ready"}
	resp, err := http.Get(ts.URL + "/ws/serviceLogs")
	if err != nil {
		t.Fatalf("failed to GET /ws/serviceLogs: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for HTTP GET on ws stub, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// WebSocket handshake test
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/ws/serviceLogs", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	key := "dGhlIHNhbXBsZSBub25jZQ=="
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", key)
	req.Header.Set("Sec-WebSocket-Version", "13")

	wsResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send websocket upgrade: %v", err)
	}
	defer wsResp.Body.Close()

	if wsResp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101 Switching Protocols, got %d", wsResp.StatusCode)
	}

	h := sha1.New()
	h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	expectedAccept := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if wsResp.Header.Get("Sec-WebSocket-Accept") != expectedAccept {
		t.Fatalf("expected Sec-WebSocket-Accept %s, got %s", expectedAccept, wsResp.Header.Get("Sec-WebSocket-Accept"))
	}
}
