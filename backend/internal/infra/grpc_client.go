package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	pb "myfi-backend/internal/generated/proto/ezistockpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Default addresses for the Python AI Service.
const (
	defaultGRPCAddr = "localhost:50051"
	defaultRESTAddr = "http://localhost:8000"
)

// GRPCClient wraps communication with the Python AI Service.
// It connects via gRPC (primary) and falls back to REST when gRPC is unavailable.
//
// Configurable via environment variables:
//   - PYTHON_AI_GRPC_ADDR: gRPC address (default: localhost:50051)
//   - PYTHON_AI_REST_ADDR: REST fallback address (default: http://localhost:8000)
//
// Requirements: 1.2 (gRPC inter-service calls with REST fallback), 1.6 (graceful degradation)
type GRPCClient struct {
	grpcAddr string
	restAddr string

	conn   *grpc.ClientConn
	connMu sync.RWMutex

	// Typed gRPC service clients (created lazily from conn).
	agentClient    pb.AgentServiceClient
	alphaClient    pb.AlphaMiningServiceClient
	feedbackClient pb.FeedbackServiceClient

	httpClient *http.Client
	logger     *slog.Logger
}

// NewGRPCClient creates a new client for the Python AI Service.
// It attempts a gRPC connection immediately but does not fail if unavailable —
// requests will fall back to REST until gRPC becomes available.
func NewGRPCClient(logger *slog.Logger) (*GRPCClient, error) {
	grpcAddr := os.Getenv("PYTHON_AI_GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = defaultGRPCAddr
	}

	restAddr := os.Getenv("PYTHON_AI_REST_ADDR")
	if restAddr == "" {
		restAddr = defaultRESTAddr
	}

	if logger == nil {
		logger = slog.Default()
	}

	c := &GRPCClient{
		grpcAddr: grpcAddr,
		restAddr: restAddr,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: logger,
	}

	// Attempt gRPC connection (non-blocking).
	if err := c.dialGRPC(); err != nil {
		logger.Warn("gRPC connection to Python AI Service failed, will use REST fallback",
			"grpcAddr", grpcAddr, "error", err)
	} else {
		logger.Info("gRPC connection to Python AI Service established", "grpcAddr", grpcAddr)
	}

	return c, nil
}

// dialGRPC establishes the gRPC connection with keepalive and timeout settings.
func (c *GRPCClient) dialGRPC() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(16*1024*1024), // 16 MB
		),
	)
	if err != nil {
		return fmt.Errorf("grpc dial: %w", err)
	}

	c.connMu.Lock()
	c.conn = conn
	c.agentClient = pb.NewAgentServiceClient(conn)
	c.alphaClient = pb.NewAlphaMiningServiceClient(conn)
	c.feedbackClient = pb.NewFeedbackServiceClient(conn)
	c.connMu.Unlock()
	return nil
}

// isGRPCAvailable checks whether the gRPC connection is in a usable state.
func (c *GRPCClient) isGRPCAvailable() bool {
	c.connMu.RLock()
	conn := c.conn
	c.connMu.RUnlock()

	if conn == nil {
		return false
	}
	state := conn.GetState()
	return state == connectivity.Ready || state == connectivity.Idle
}

// Close shuts down the gRPC connection.
func (c *GRPCClient) Close() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Conn returns the underlying gRPC connection.
// Returns nil if gRPC is not connected.
func (c *GRPCClient) Conn() *grpc.ClientConn {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.conn
}

// AgentClient returns the AgentService gRPC client. Returns nil if not connected.
func (c *GRPCClient) AgentClient() pb.AgentServiceClient {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.agentClient
}

// AlphaClient returns the AlphaMiningService gRPC client. Returns nil if not connected.
func (c *GRPCClient) AlphaClient() pb.AlphaMiningServiceClient {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.alphaClient
}

// FeedbackClient returns the FeedbackService gRPC client. Returns nil if not connected.
func (c *GRPCClient) FeedbackClient() pb.FeedbackServiceClient {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.feedbackClient
}

// ---------------------------------------------------------------------------
// Health Check
// ---------------------------------------------------------------------------

// HealthStatus represents the health of the Python AI Service.
type HealthStatus struct {
	GRPC      bool   `json:"grpc"`
	REST      bool   `json:"rest"`
	Available bool   `json:"available"`
	Mode      string `json:"mode"` // "grpc", "rest", or "unavailable"
}

// HealthCheck probes the Python AI Service via gRPC and REST.
func (c *GRPCClient) HealthCheck(ctx context.Context) HealthStatus {
	status := HealthStatus{}

	// Check gRPC.
	status.GRPC = c.isGRPCAvailable()

	// Check REST health endpoint.
	status.REST = c.restHealthCheck(ctx)

	// Determine overall availability.
	switch {
	case status.GRPC:
		status.Available = true
		status.Mode = "grpc"
	case status.REST:
		status.Available = true
		status.Mode = "rest"
	default:
		status.Mode = "unavailable"
	}

	return status
}

func (c *GRPCClient) restHealthCheck(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.restAddr+"/health", nil)
	if err != nil {
		return false
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ---------------------------------------------------------------------------
// AgentService Methods — gRPC primary, REST fallback
// ---------------------------------------------------------------------------

// AnalyzeStock sends a stock analysis request to the Python AI Service.
func (c *GRPCClient) AnalyzeStock(ctx context.Context, req *pb.AnalyzeStockRequest) (*pb.AnalyzeStockResponse, error) {
	if c.isGRPCAvailable() {
		client := c.AgentClient()
		if client != nil {
			resp, err := client.AnalyzeStock(ctx, req)
			if err == nil {
				return resp, nil
			}
			c.logger.Warn("gRPC AnalyzeStock failed, falling back to REST", "error", err)
		}
	}

	return c.analyzeStockREST(ctx, req)
}

func (c *GRPCClient) analyzeStockREST(ctx context.Context, req *pb.AnalyzeStockRequest) (*pb.AnalyzeStockResponse, error) {
	body := map[string]any{"symbol": req.Symbol}
	var resp pb.AnalyzeStockResponse
	if err := c.postJSON(ctx, "/api/analyze", body, &resp); err != nil {
		return nil, fmt.Errorf("REST AnalyzeStock: %w", err)
	}
	return &resp, nil
}

// Chat sends a chat message to the Python AI Service.
func (c *GRPCClient) Chat(ctx context.Context, req *pb.ChatRequest) (*pb.ChatResponse, error) {
	if c.isGRPCAvailable() {
		client := c.AgentClient()
		if client != nil {
			resp, err := client.Chat(ctx, req)
			if err == nil {
				return resp, nil
			}
			c.logger.Warn("gRPC Chat failed, falling back to REST", "error", err)
		}
	}

	return c.chatREST(ctx, req)
}

func (c *GRPCClient) chatREST(ctx context.Context, req *pb.ChatRequest) (*pb.ChatResponse, error) {
	body := map[string]any{"message": req.Message, "user_id": req.UserId}
	var resp pb.ChatResponse
	if err := c.postJSON(ctx, "/api/chat", body, &resp); err != nil {
		return nil, fmt.Errorf("REST Chat: %w", err)
	}
	return &resp, nil
}

// GenerateInvestmentIdeas fetches proactive investment ideas from the Python AI Service.
func (c *GRPCClient) GenerateInvestmentIdeas(ctx context.Context, req *pb.IdeaRequest) (*pb.IdeaResponse, error) {
	if c.isGRPCAvailable() {
		client := c.AgentClient()
		if client != nil {
			resp, err := client.GenerateInvestmentIdeas(ctx, req)
			if err == nil {
				return resp, nil
			}
			c.logger.Warn("gRPC GenerateInvestmentIdeas failed, falling back to REST", "error", err)
		}
	}

	return c.generateIdeasREST(ctx, req)
}

func (c *GRPCClient) generateIdeasREST(ctx context.Context, req *pb.IdeaRequest) (*pb.IdeaResponse, error) {
	body := map[string]any{
		"user_id":   req.UserId,
		"max_ideas": req.MaxIdeas,
	}
	var resp pb.IdeaResponse
	if err := c.postJSON(ctx, "/api/ideas", body, &resp); err != nil {
		return nil, fmt.Errorf("REST GenerateInvestmentIdeas: %w", err)
	}
	return &resp, nil
}

// GetHotTopics fetches trending stocks and market events from the Python AI Service.
func (c *GRPCClient) GetHotTopics(ctx context.Context, req *pb.HotTopicsRequest) (*pb.HotTopicsResponse, error) {
	if c.isGRPCAvailable() {
		client := c.AgentClient()
		if client != nil {
			resp, err := client.GetHotTopics(ctx, req)
			if err == nil {
				return resp, nil
			}
			c.logger.Warn("gRPC GetHotTopics failed, falling back to REST", "error", err)
		}
	}

	return c.getHotTopicsREST(ctx, req)
}

func (c *GRPCClient) getHotTopicsREST(ctx context.Context, req *pb.HotTopicsRequest) (*pb.HotTopicsResponse, error) {
	body := map[string]any{"limit": req.Limit, "market": req.Market}
	var resp pb.HotTopicsResponse
	if err := c.postJSON(ctx, "/api/hot-topics", body, &resp); err != nil {
		return nil, fmt.Errorf("REST GetHotTopics: %w", err)
	}
	return &resp, nil
}

// ---------------------------------------------------------------------------
// AlphaMiningService Methods — gRPC primary, REST fallback
// ---------------------------------------------------------------------------

// GetRanking fetches stock rankings from the Python AI Service.
func (c *GRPCClient) GetRanking(ctx context.Context, req *pb.RankingRequest) (*pb.RankingResponse, error) {
	if c.isGRPCAvailable() {
		client := c.AlphaClient()
		if client != nil {
			resp, err := client.GetRanking(ctx, req)
			if err == nil {
				return resp, nil
			}
			c.logger.Warn("gRPC GetRanking failed, falling back to REST", "error", err)
		}
	}

	return c.getRankingREST(ctx, req)
}

func (c *GRPCClient) getRankingREST(ctx context.Context, req *pb.RankingRequest) (*pb.RankingResponse, error) {
	body := map[string]any{
		"universe":      req.Universe,
		"factor_groups": req.FactorGroups,
		"top_n":         req.TopN,
	}
	var resp pb.RankingResponse
	if err := c.postJSON(ctx, "/api/ranking", body, &resp); err != nil {
		return nil, fmt.Errorf("REST GetRanking: %w", err)
	}
	return &resp, nil
}

// RunBacktest executes a walk-forward backtest via the Python AI Service.
func (c *GRPCClient) RunBacktest(ctx context.Context, req *pb.BacktestRequest) (*pb.BacktestResponse, error) {
	if c.isGRPCAvailable() {
		client := c.AlphaClient()
		if client != nil {
			resp, err := client.RunBacktest(ctx, req)
			if err == nil {
				return resp, nil
			}
			c.logger.Warn("gRPC RunBacktest failed, falling back to REST", "error", err)
		}
	}

	return c.runBacktestREST(ctx, req)
}

func (c *GRPCClient) runBacktestREST(ctx context.Context, req *pb.BacktestRequest) (*pb.BacktestResponse, error) {
	body := map[string]any{
		"universe":      req.Universe,
		"factor_groups": req.FactorGroups,
		"start_date":    req.StartDate,
		"end_date":      req.EndDate,
		"top_n":         req.TopN,
	}
	var resp pb.BacktestResponse
	if err := c.postJSON(ctx, "/api/ranking/backtest", body, &resp); err != nil {
		return nil, fmt.Errorf("REST RunBacktest: %w", err)
	}
	return &resp, nil
}

// GetRegime returns the current market regime classification.
func (c *GRPCClient) GetRegime(ctx context.Context, req *pb.RegimeRequest) (*pb.RegimeResponse, error) {
	if c.isGRPCAvailable() {
		client := c.AlphaClient()
		if client != nil {
			resp, err := client.GetRegime(ctx, req)
			if err == nil {
				return resp, nil
			}
			c.logger.Warn("gRPC GetRegime failed, falling back to REST", "error", err)
		}
	}

	return c.getRegimeREST(ctx)
}

func (c *GRPCClient) getRegimeREST(ctx context.Context) (*pb.RegimeResponse, error) {
	body := map[string]any{}
	var resp pb.RegimeResponse
	if err := c.postJSON(ctx, "/api/regime", body, &resp); err != nil {
		return nil, fmt.Errorf("REST GetRegime: %w", err)
	}
	return &resp, nil
}

// ---------------------------------------------------------------------------
// FeedbackService Methods — gRPC primary, REST fallback
// ---------------------------------------------------------------------------

// GetAgentAccuracy returns per-agent accuracy scores from the Python AI Service.
func (c *GRPCClient) GetAgentAccuracy(ctx context.Context, req *pb.AccuracyRequest) (*pb.AccuracyResponse, error) {
	if c.isGRPCAvailable() {
		client := c.FeedbackClient()
		if client != nil {
			resp, err := client.GetAgentAccuracy(ctx, req)
			if err == nil {
				return resp, nil
			}
			c.logger.Warn("gRPC GetAgentAccuracy failed, falling back to REST", "error", err)
		}
	}

	return c.getAgentAccuracyREST(ctx, req)
}

func (c *GRPCClient) getAgentAccuracyREST(ctx context.Context, req *pb.AccuracyRequest) (*pb.AccuracyResponse, error) {
	body := map[string]any{"agent_name": req.AgentName, "period": req.Period}
	var resp pb.AccuracyResponse
	if err := c.postJSON(ctx, "/api/feedback/accuracy", body, &resp); err != nil {
		return nil, fmt.Errorf("REST GetAgentAccuracy: %w", err)
	}
	return &resp, nil
}

// GetModelPerformance returns Alpha Mining model performance metrics.
func (c *GRPCClient) GetModelPerformance(ctx context.Context, req *pb.ModelPerfRequest) (*pb.ModelPerfResponse, error) {
	if c.isGRPCAvailable() {
		client := c.FeedbackClient()
		if client != nil {
			resp, err := client.GetModelPerformance(ctx, req)
			if err == nil {
				return resp, nil
			}
			c.logger.Warn("gRPC GetModelPerformance failed, falling back to REST", "error", err)
		}
	}

	return c.getModelPerformanceREST(ctx, req)
}

func (c *GRPCClient) getModelPerformanceREST(ctx context.Context, req *pb.ModelPerfRequest) (*pb.ModelPerfResponse, error) {
	body := map[string]any{"model_name": req.ModelName, "period": req.Period}
	var resp pb.ModelPerfResponse
	if err := c.postJSON(ctx, "/api/feedback/model-performance", body, &resp); err != nil {
		return nil, fmt.Errorf("REST GetModelPerformance: %w", err)
	}
	return &resp, nil
}

// ---------------------------------------------------------------------------
// REST helpers
// ---------------------------------------------------------------------------

// postJSON sends a POST request with JSON body to the REST fallback endpoint.
func (c *GRPCClient) postJSON(ctx context.Context, path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := c.restAddr + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("unexpected status %d from %s: %s", resp.StatusCode, url, string(respBody))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response from %s: %w", url, err)
		}
	}
	return nil
}
