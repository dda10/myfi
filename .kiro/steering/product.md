# Product Overview

EziStock is a unified finance platform for Vietnamese stock market analysis and portfolio management.

## Core Features

- Real-time stock quotes and market data (Vietnamese exchanges: HOSE, HNX, UPCOM)
- Interactive candlestick charts with volume indicators and technical overlays
- Stock screener with fundamental metrics (P/E, P/B, EV/EBITDA, ROE, market cap) — delegates to vnstock-go Screen
- Liquidity scoring (0-100) with real bid-ask spread from PriceBoard and free-float estimation
- Portfolio tracking, performance (TWR, XIRR), risk metrics (Sharpe, VaR), and corporate action handling
- AI-powered multi-agent analysis (technical, news, investment advisor, strategy builder)
- Macro dashboard: interbank rates, bond yields, VCB FX rates, SJC/BTMC gold prices
- AI stock ranking with backtesting
- Knowledge base with pattern accuracy tracking
- Missions (price alerts, scheduled triggers) with trading-hours-aware scheduling
- Mutual fund data: listings, holdings, NAV history, allocation (via FMARKET connector)
- News aggregation and sentiment analysis

## Target Market

Vietnamese retail investors and traders seeking comprehensive market analysis tools.

## Data Sources

- vnstock-go v2 multi-connector architecture:
  - VCI (primary), KBS, VND, ENTRADE, CAFEF, VND_FINFO for stock data
  - GOLD connector for SJC/BTMC gold prices
  - MSN connector for world market data
  - FMARKET connector for fund data
- VCBExchangeRate for official Vietcombank FX rates
- Automatic multi-level failover with circuit breaker protection
- Python AI Service (gRPC) for multi-agent analysis, alpha mining, feedback loop
