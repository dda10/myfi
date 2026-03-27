package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"myfi-backend/internal/model"

	"github.com/tmc/langchaingo/llms"
)

// ---------------------------------------------------------------------------
// Price_Agent — fetches current and historical prices for the AI system
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 9.1  Fetch current price via Data_Source_Router (VCI/KBS for stocks,
//          CoinGecko for crypto, Doji for gold)
//   - 9.2  Fetch OHLCV history for specified time range and interval
//   - 9.3  Return structured PriceAgentResponse (price, change, volume, source,
//          historical data)
//   - 9.4  Failover handled transparently by Data_Source_Router / PriceService
//   - 9.5  All prices formatted in VND

// PriceAgent is the AI sub-agent responsible for fetching current and
// historical price data for any supported asset type. It delegates data
// retrieval to PriceService (which uses DataSourceRouter for failover) and
// CryptoService / GoldService for non-stock assets.
type PriceAgent struct {
	priceService  *PriceService
	cryptoService *CryptoService
	fxService     *FXService
}

// NewPriceAgent creates a PriceAgent with the required service dependencies.
func NewPriceAgent(priceService *PriceService, cryptoService *CryptoService, fxService *FXService) *PriceAgent {
	return &PriceAgent{
		priceService:  priceService,
		cryptoService: cryptoService,
		fxService:     fxService,
	}
}

// Name returns the agent identifier used by the orchestrator.
func (a *PriceAgent) Name() string { return "Price_Agent" }

// Execute fetches price data for the symbols in the query intent and returns
// an AgentMessage containing []PriceAgentResponse in the payload.
//
// The LLM parameter is accepted to satisfy the SubAgent interface but is not
// used — the Price_Agent is a pure data-fetching agent.
func (a *PriceAgent) Execute(ctx context.Context, intent model.QueryIntent, _ llms.Model) (*model.AgentMessage, error) {
	if len(intent.Symbols) == 0 {
		return nil, fmt.Errorf("Price_Agent: no symbols provided in query intent")
	}

	assetType := a.resolveAssetType(intent)

	var responses []model.PriceAgentResponse
	for _, symbol := range intent.Symbols {
		resp, err := a.fetchForSymbol(ctx, symbol, assetType)
		if err != nil {
			log.Printf("[Price_Agent] failed to fetch %s (%s): %v", symbol, assetType, err)
			continue // partial failure tolerance — skip this symbol
		}
		responses = append(responses, *resp)
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("Price_Agent: failed to fetch prices for all requested symbols")
	}

	// Pack responses into AgentMessage payload
	payload := make(map[string]interface{})
	payload["prices"] = responses

	msg := &model.AgentMessage{
		AgentName:   a.Name(),
		PayloadType: "price_data",
		Payload:     payload,
		Timestamp:   time.Now(),
	}

	log.Printf("[Price_Agent] successfully fetched prices for %d/%d symbols", len(responses), len(intent.Symbols))
	return msg, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// resolveAssetType determines the asset type from the query intent. If the
// intent explicitly specifies asset types, the first one is used; otherwise
// it defaults to VNStock.
func (a *PriceAgent) resolveAssetType(intent model.QueryIntent) model.AssetType {
	for _, at := range intent.AssetTypes {
		switch model.AssetType(at) {
		case model.Crypto:
			return model.Crypto
		case model.Gold:
			return model.Gold
		case model.VNStock:
			return model.VNStock
		}
	}
	return model.VNStock // default
}

// fetchForSymbol fetches current price + recent historical data for a single
// symbol and returns a PriceAgentResponse with all prices in VND.
func (a *PriceAgent) fetchForSymbol(ctx context.Context, symbol string, assetType model.AssetType) (*model.PriceAgentResponse, error) {
	switch assetType {
	case model.VNStock:
		return a.fetchStock(ctx, symbol)
	case model.Crypto:
		return a.fetchCrypto(ctx, symbol)
	case model.Gold:
		return a.fetchGold(ctx, symbol)
	default:
		return nil, fmt.Errorf("unsupported asset type: %s", assetType)
	}
}

// fetchStock fetches VN stock price and 90-day history via PriceService.
// Stock prices from VCI/KBS are already denominated in VND.
func (a *PriceAgent) fetchStock(ctx context.Context, symbol string) (*model.PriceAgentResponse, error) {
	// Fetch current quote
	quotes, err := a.priceService.GetQuotes(ctx, []string{symbol}, model.VNStock)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stock quote for %s: %w", symbol, err)
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quote data returned for %s", symbol)
	}

	quote := quotes[0]

	resp := &model.PriceAgentResponse{
		Symbol:        quote.Symbol,
		CurrentPrice:  quote.Price,  // already VND
		Change:        quote.Change, // already VND
		ChangePercent: quote.ChangePercent,
		Volume:        quote.Volume,
		Source:        quote.Source,
	}

	// Fetch 90-day historical data for AI context
	history, err := a.fetchHistory(ctx, symbol, 90)
	if err != nil {
		log.Printf("[Price_Agent] historical data unavailable for %s: %v", symbol, err)
		// Non-fatal — return current price without history
	} else {
		resp.HistoricalData = history
	}

	return resp, nil
}

// fetchCrypto fetches crypto price via CryptoService and converts to VND.
func (a *PriceAgent) fetchCrypto(ctx context.Context, symbol string) (*model.PriceAgentResponse, error) {
	if a.cryptoService == nil {
		return nil, fmt.Errorf("CryptoService not available")
	}

	cryptoPrice, err := a.cryptoService.GetCryptoPriceBySymbol(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crypto price for %s: %w", symbol, err)
	}

	// Convert USD price to VND (Requirement 9.5)
	priceVND := cryptoPrice.PriceVND
	if priceVND == 0 && cryptoPrice.PriceUSD > 0 && a.fxService != nil {
		converted, fxErr := a.fxService.ConvertUSDToVND(cryptoPrice.PriceUSD)
		if fxErr == nil {
			priceVND = converted
		} else {
			// Fallback: use hardcoded rate
			priceVND = cryptoPrice.PriceUSD * model.FallbackUSDVND
		}
	}

	changeVND := 0.0
	if cryptoPrice.Change24h != 0 && a.fxService != nil {
		converted, fxErr := a.fxService.ConvertUSDToVND(cryptoPrice.Change24h)
		if fxErr == nil {
			changeVND = converted
		} else {
			changeVND = cryptoPrice.Change24h * model.FallbackUSDVND
		}
	}

	return &model.PriceAgentResponse{
		Symbol:        cryptoPrice.Symbol,
		CurrentPrice:  priceVND,
		Change:        changeVND,
		ChangePercent: cryptoPrice.PercentChange24h,
		Volume:        int64(cryptoPrice.Volume24h),
		Source:        cryptoPrice.Source,
	}, nil
}

// fetchGold fetches gold price via PriceService (which delegates to GoldService).
// Gold prices from Doji/SJC are already in VND.
func (a *PriceAgent) fetchGold(ctx context.Context, symbol string) (*model.PriceAgentResponse, error) {
	quotes, err := a.priceService.GetQuotes(ctx, []string{symbol}, model.Gold)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gold price for %s: %w", symbol, err)
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no gold price data returned for %s", symbol)
	}

	quote := quotes[0]

	return &model.PriceAgentResponse{
		Symbol:        quote.Symbol,
		CurrentPrice:  quote.Price,  // already VND
		Change:        quote.Change, // already VND
		ChangePercent: quote.ChangePercent,
		Volume:        quote.Volume,
		Source:        quote.Source,
	}, nil
}

// fetchHistory fetches OHLCV bars for the last N days via PriceService.
func (a *PriceAgent) fetchHistory(ctx context.Context, symbol string, days int) ([]model.OHLCVBar, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)

	bars, err := a.priceService.GetHistoricalData(ctx, symbol, start, end, "1D")
	if err != nil {
		return nil, err
	}
	return bars, nil
}
