package platform_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"myfi-backend/internal/domain/agent"
	"myfi-backend/internal/domain/analyst"
	domainAuth "myfi-backend/internal/domain/auth"
	"myfi-backend/internal/domain/consensus"
	"myfi-backend/internal/domain/knowledge"
	"myfi-backend/internal/domain/market"
	"myfi-backend/internal/domain/mission"
	"myfi-backend/internal/domain/notification"
	"myfi-backend/internal/domain/portfolio"
	"myfi-backend/internal/domain/ranking"
	"myfi-backend/internal/domain/research"
	"myfi-backend/internal/domain/screener"
	"myfi-backend/internal/domain/sentiment"
	"myfi-backend/internal/domain/watchlist"
	"myfi-backend/internal/infra"
	"myfi-backend/internal/platform"

	"github.com/gin-gonic/gin"
)

// testServer creates a test HTTP server with the full router wired up.
// Uses in-memory/mock dependencies — no real DB or external services.
func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cache := infra.NewCache()

	cfg := platform.Config{
		Port:           0,
		Env:            "test",
		FrontendOrigin: "http://localhost:3000",
		IPRateLimit:    1000,
		UserRateLimit:  1000,
	}

	// Build minimal domain handlers with nil/mock dependencies.
	// Handlers that need a DB will return appropriate errors.
	dh := &platform.DomainHandlers{
		Market: &market.Handlers{
			PriceService:  market.NewPriceService(nil, cache),
			SectorService: market.NewSectorService(nil, cache),
			MacroService:  market.NewMacroService(cache, nil),
			SearchService: market.NewSearchService(nil, cache),
		},
		Portfolio:    &portfolio.Handlers{},
		Screener:     &screener.Handlers{},
		Watchlist:    &watchlist.Handlers{},
		Auth:         &domainAuth.Handlers{AuthService: domainAuth.NewAuthService(nil, domainAuth.DefaultAuthConfig())},
		Agent:        &agent.Handlers{},
		Ranking:      &ranking.Handlers{Tracker: ranking.NewRecommendationTracker(nil)},
		Mission:      &mission.Handlers{},
		Notification: &notification.Handlers{},
		Knowledge:    &knowledge.Handlers{},
		Analyst:      &analyst.Handlers{},
		Research:     &research.Handlers{},
		Consensus:    &consensus.Handlers{},
		Sentiment:    &sentiment.Handlers{},
	}

	engine := platform.NewRouter(cfg, dh)
	return httptest.NewServer(engine)
}

// --- Health Endpoint Tests ---

func TestHealthEndpoint(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET /api/health failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", body["status"])
	}
}

func TestHealthzEndpoint(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/healthz")
	if err != nil {
		t.Fatalf("GET /api/healthz failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// --- Auth Endpoint Tests ---

func TestLoginWithoutCredentials(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/auth/login", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST /api/auth/login failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Error("expected non-200 for empty login, got 200")
	}
}

// --- Protected Route Tests (no JWT → 401) ---

func TestProtectedRouteWithoutJWT(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	endpoints := []string{
		"/api/market/quote",
		"/api/portfolio/holdings",
		"/api/watchlists",
		"/api/screener/presets",
		"/api/notifications",
		"/api/missions",
		"/api/ranking/factors",
		"/api/ideas",
	}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			resp, err := http.Get(srv.URL + ep)
			if err != nil {
				t.Fatalf("GET %s failed: %v", ep, err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("GET %s: expected 401, got %d", ep, resp.StatusCode)
			}
		})
	}
}

// --- OpenAPI Documentation Tests ---

func TestSwaggerJSONEndpoint(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/docs/swagger.json")
	if err != nil {
		t.Fatalf("GET /api/docs/swagger.json failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var spec map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&spec); err != nil {
		t.Fatalf("failed to decode swagger.json: %v", err)
	}

	if spec["openapi"] != "3.0.3" {
		t.Errorf("expected openapi=3.0.3, got %v", spec["openapi"])
	}

	info, ok := spec["info"].(map[string]any)
	if !ok || info["title"] != "EziStock API" {
		t.Errorf("expected info.title=EziStock API, got %v", info)
	}
}

func TestSwaggerUIEndpoint(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/docs")
	if err != nil {
		t.Fatalf("GET /api/docs failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type text/html, got %s", ct)
	}
}

// --- Prometheus Metrics Tests ---

func TestMetricsEndpoint(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	// Make a request first to generate some metrics.
	if warmup, err := http.Get(srv.URL + "/api/health"); err == nil {
		warmup.Body.Close()
	}

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected Prometheus text/plain content type, got %s", ct)
	}
}

// --- AI Service Degradation Tests ---

func TestChatWithoutJWT(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/chat", "application/json", strings.NewReader(`{"message":"test"}`))
	if err != nil {
		t.Fatalf("POST /api/chat failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without JWT, got %d", resp.StatusCode)
	}
}

// --- CORS Tests ---

func TestCORSHeaders(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	req, _ := http.NewRequest("OPTIONS", srv.URL+"/api/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS /api/health failed: %v", err)
	}
	defer resp.Body.Close()

	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:3000" {
		t.Errorf("expected CORS origin http://localhost:3000, got %s", origin)
	}
}

// --- Security Headers Tests ---

func TestSecurityHeaders(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET /api/health failed: %v", err)
	}
	defer resp.Body.Close()

	if v := resp.Header.Get("X-Content-Type-Options"); v != "nosniff" {
		t.Errorf("expected X-Content-Type-Options=nosniff, got %s", v)
	}
	if v := resp.Header.Get("X-Frame-Options"); v != "DENY" {
		t.Errorf("expected X-Frame-Options=DENY, got %s", v)
	}
}
