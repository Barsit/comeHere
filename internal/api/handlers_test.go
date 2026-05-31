package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/db/comehere/internal/certmgr"
	"github.com/db/comehere/internal/config"
	"github.com/db/comehere/internal/hostsmgr"
	"github.com/db/comehere/internal/proxy"
)

func newTestServer(t *testing.T) (*Server, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := config.DefaultConfig()
	hm := hostsmgr.New()
	hm.Path = t.TempDir() + "\\hosts"
	os.WriteFile(hm.Path, []byte("127.0.0.1 localhost\n"), 0644)

	cm := certmgr.New(t.TempDir())
	cm.Dir = t.TempDir() + "\\certs"

	px := proxy.New(nil, 8443)

	s := New(cfg, hm, cm, px)
	r := gin.New()
	s.RegisterRoutes(r)
	return s, r
}

func performRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---- Rule CRUD Tests ----

func TestListRules_Empty(t *testing.T) {
	_, r := newTestServer(t)
	w := performRequest(r, "GET", "/api/rules", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	rules := resp["rules"].([]interface{})
	if len(rules) != 0 {
		t.Errorf("expected empty rules, got %d", len(rules))
	}
}

func TestCreateRule_Success(t *testing.T) {
	s, r := newTestServer(t)
	body := `{"source":"api.test.com","source_port":443,"target":"localhost:8080","description":"test"}`
	w := performRequest(r, "POST", "/api/rules", body)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if len(s.Config.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(s.Config.Rules))
	}
	if s.Config.Rules[0].Source != "api.test.com" {
		t.Errorf("expected api.test.com, got %s", s.Config.Rules[0].Source)
	}
	if s.Config.Rules[0].Enabled != false {
		t.Error("new rule should be disabled by default")
	}
	if s.Config.Rules[0].ID == "" {
		t.Error("rule ID should not be empty")
	}
}

func TestCreateRule_DuplicateDomain(t *testing.T) {
	_, r := newTestServer(t)
	body := `{"source":"api.test.com","target":"localhost:8080"}`
	performRequest(r, "POST", "/api/rules", body)
	w := performRequest(r, "POST", "/api/rules", body)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateRule_InvalidJSON(t *testing.T) {
	_, r := newTestServer(t)
	w := performRequest(r, "POST", "/api/rules", "not-json")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetRules_WithItems(t *testing.T) {
	_, r := newTestServer(t)
	performRequest(r, "POST", "/api/rules", `{"source":"a.com","target":"localhost:1"}`)
	performRequest(r, "POST", "/api/rules", `{"source":"b.com","target":"localhost:2"}`)

	w := performRequest(r, "GET", "/api/rules", "")
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	rules := resp["rules"].([]interface{})
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}

func TestUpdateRule_Success(t *testing.T) {
	s, r := newTestServer(t)
	performRequest(r, "POST", "/api/rules", `{"source":"a.com","target":"localhost:1"}`)
	id := s.Config.Rules[0].ID

	body := `{"source":"a.com","target":"localhost:9999","description":"updated"}`
	w := performRequest(r, "PUT", "/api/rules/"+id, body)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if s.Config.Rules[0].Target != "localhost:9999" {
		t.Errorf("expected localhost:9999, got %s", s.Config.Rules[0].Target)
	}
}

func TestUpdateRule_NotFound(t *testing.T) {
	_, r := newTestServer(t)
	w := performRequest(r, "PUT", "/api/rules/nonexistent", `{"source":"x.com","target":"x"}`)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdateRule_WhenEnabled(t *testing.T) {
	s, r := newTestServer(t)
	performRequest(r, "POST", "/api/rules", `{"source":"a.com","target":"localhost:1"}`)
	id := s.Config.Rules[0].ID
	// Enable it
	performRequest(r, "POST", "/api/rules/"+id+"/enable", "")

	w := performRequest(r, "PUT", "/api/rules/"+id, `{"source":"a.com","target":"changed"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when editing enabled rule, got %d", w.Code)
	}
}

func TestDeleteRule_Success(t *testing.T) {
	s, r := newTestServer(t)
	performRequest(r, "POST", "/api/rules", `{"source":"a.com","target":"localhost:1"}`)
	id := s.Config.Rules[0].ID

	w := performRequest(r, "DELETE", "/api/rules/"+id, "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if len(s.Config.Rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(s.Config.Rules))
	}
}

func TestDeleteRule_NotFound(t *testing.T) {
	_, r := newTestServer(t)
	w := performRequest(r, "DELETE", "/api/rules/nonexistent", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDeleteRule_WhenEnabled(t *testing.T) {
	s, r := newTestServer(t)
	performRequest(r, "POST", "/api/rules", `{"source":"a.com","target":"localhost:1"}`)
	id := s.Config.Rules[0].ID
	performRequest(r, "POST", "/api/rules/"+id+"/enable", "")

	w := performRequest(r, "DELETE", "/api/rules/"+id, "")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when deleting enabled rule, got %d", w.Code)
	}
}

// ---- Enable/Disable Tests ----

func TestEnableRule(t *testing.T) {
	s, r := newTestServer(t)
	performRequest(r, "POST", "/api/rules", `{"source":"api.test.com","target":"localhost:8080"}`)
	id := s.Config.Rules[0].ID

	w := performRequest(r, "POST", "/api/rules/"+id+"/enable", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !s.Config.Rules[0].Enabled {
		t.Error("rule should be enabled after enable")
	}
	// Should have been added to proxy
	if !s.Proxy.HasRule("api.test.com") {
		t.Error("rule should be in proxy engine after enable")
	}
	// hosts file should have the entry
	has, _ := s.HostsMgr.HasDomain("api.test.com")
	if !has {
		t.Error("domain should be in hosts after enable")
	}
}

func TestEnableRule_AlreadyEnabled(t *testing.T) {
	s, r := newTestServer(t)
	performRequest(r, "POST", "/api/rules", `{"source":"a.com","target":"localhost:1"}`)
	id := s.Config.Rules[0].ID
	performRequest(r, "POST", "/api/rules/"+id+"/enable", "")
	w := performRequest(r, "POST", "/api/rules/"+id+"/enable", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 even if already enabled, got %d", w.Code)
	}
}

func TestEnableRule_NotFound(t *testing.T) {
	_, r := newTestServer(t)
	w := performRequest(r, "POST", "/api/rules/nonexistent/enable", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDisableRule(t *testing.T) {
	s, r := newTestServer(t)
	performRequest(r, "POST", "/api/rules", `{"source":"api.test.com","target":"localhost:8080"}`)
	id := s.Config.Rules[0].ID
	performRequest(r, "POST", "/api/rules/"+id+"/enable", "")

	w := performRequest(r, "POST", "/api/rules/"+id+"/disable", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if s.Config.Rules[0].Enabled {
		t.Error("rule should be disabled after disable")
	}
	if s.Proxy.HasRule("api.test.com") {
		t.Error("rule should be removed from proxy engine")
	}
}

func TestDisableRule_AlreadyDisabled(t *testing.T) {
	s, r := newTestServer(t)
	performRequest(r, "POST", "/api/rules", `{"source":"a.com","target":"localhost:1"}`)
	id := s.Config.Rules[0].ID
	w := performRequest(r, "POST", "/api/rules/"+id+"/disable", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 even if already disabled, got %d", w.Code)
	}
}

func TestDisableRule_NotFound(t *testing.T) {
	_, r := newTestServer(t)
	w := performRequest(r, "POST", "/api/rules/nonexistent/disable", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---- Status Tests ----

func TestGetStatus_Empty(t *testing.T) {
	_, r := newTestServer(t)
	w := performRequest(r, "GET", "/api/status", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["rules_total"].(float64) != 0 {
		t.Errorf("expected 0, got %v", resp["rules_total"])
	}
	if resp["rules_enabled"].(float64) != 0 {
		t.Errorf("expected 0, got %v", resp["rules_enabled"])
	}
}

func TestGetStatus_WithRules(t *testing.T) {
	s, r := newTestServer(t)
	performRequest(r, "POST", "/api/rules", `{"source":"a.com","target":"localhost:1"}`)
	performRequest(r, "POST", "/api/rules", `{"source":"b.com","target":"localhost:2"}`)
	id := s.Config.Rules[0].ID
	performRequest(r, "POST", "/api/rules/"+id+"/enable", "")

	w := performRequest(r, "GET", "/api/status", "")
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["rules_total"].(float64) != 2 {
		t.Errorf("expected 2 total, got %v", resp["rules_total"])
	}
	if resp["rules_enabled"].(float64) != 1 {
		t.Errorf("expected 1 enabled, got %v", resp["rules_enabled"])
	}
	if resp["listen_port"].(float64) != 443 {
		t.Errorf("expected 443, got %v", resp["listen_port"])
	}
	if resp["admin_port"].(float64) != 8848 {
		t.Errorf("expected 8848, got %v", resp["admin_port"])
	}
}

// ---- Cleanup Tests ----

func TestCleanupHosts(t *testing.T) {
	s, r := newTestServer(t)
	// Add and enable a rule to create hosts entry
	performRequest(r, "POST", "/api/rules", `{"source":"api.test.com","target":"localhost:8080"}`)
	id := s.Config.Rules[0].ID
	performRequest(r, "POST", "/api/rules/"+id+"/enable", "")

	// Cleanup
	w := performRequest(r, "POST", "/api/cleanup", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify hosts entry removed
	has, _ := s.HostsMgr.HasDomain("api.test.com")
	if has {
		t.Error("domain should be removed after cleanup")
	}
}

// ---- Cert Status Tests ----

func TestCertStatus_NotInstalled(t *testing.T) {
	_, r := newTestServer(t)
	w := performRequest(r, "GET", "/api/cert/status", "")
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["ca_installed"].(bool) != false {
		t.Errorf("expected false when CA not installed, got %v", resp["ca_installed"])
	}
}

func TestCertStatus_Installed(t *testing.T) {
	s, r := newTestServer(t)
	// Create the CA certificate
	if err := s.CertMgr.EnsureCA(365); err != nil {
		t.Fatal(err)
	}
	w := performRequest(r, "GET", "/api/cert/status", "")
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["ca_installed"].(bool) != true {
		t.Errorf("expected true when CA installed, got %v", resp["ca_installed"])
	}
}

// ---- Server Struct Tests ----

func TestNewServer(t *testing.T) {
	cfg := config.DefaultConfig()
	hm := hostsmgr.New()
	cm := certmgr.New(t.TempDir())
	px := proxy.New(nil, 443)

	s := New(cfg, hm, cm, px)
	if s.Config != cfg {
		t.Error("Config should match")
	}
	if s.HostsMgr != hm {
		t.Error("HostsMgr should match")
	}
	if s.CertMgr != cm {
		t.Error("CertMgr should match")
	}
	if s.Proxy != px {
		t.Error("Proxy should match")
	}
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.DefaultConfig()
	hm := hostsmgr.New()
	hm.Path = t.TempDir() + "\\hosts"
	os.WriteFile(hm.Path, []byte("127.0.0.1 localhost\n"), 0644)
	cm := certmgr.New(t.TempDir())
	px := proxy.New(nil, 443)

	s := New(cfg, hm, cm, px)
	r := gin.New()
	s.RegisterRoutes(r)

	// Verify all routes are registered
	routes := r.Routes()
	expectedPaths := []string{
		"/api/rules",
		"/api/status",
		"/api/cleanup",
		"/api/cert/status",
	}
	for _, path := range expectedPaths {
		found := false
		for _, route := range routes {
			if route.Path == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("route %s not registered", path)
		}
	}
}
