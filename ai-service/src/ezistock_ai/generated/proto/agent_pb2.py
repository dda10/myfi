"""Generated protobuf message classes for ezistock/agent.proto.

Hand-written stubs matching the proto definitions. These classes use simple
__init__ / __repr__ patterns so that gRPC servicers can construct responses
without requiring a full protoc toolchain at dev time.

When a real protoc build is available, replace this file with the generated output.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Dict, List, Optional


# ---------------------------------------------------------------------------
# Shared / nested messages
# ---------------------------------------------------------------------------


@dataclass
class OHLCV:
    date: str = ""
    open: float = 0.0
    high: float = 0.0
    low: float = 0.0
    close: float = 0.0
    volume: int = 0


@dataclass
class FinancialMetric:
    name: str = ""
    value: float = 0.0
    period: str = ""


@dataclass
class NewsArticle:
    title: str = ""
    url: str = ""
    source: str = ""
    published_at: str = ""
    summary: str = ""


@dataclass
class MarketData:
    symbol: str = ""
    ohlcv: List[OHLCV] = field(default_factory=list)
    financials: List[FinancialMetric] = field(default_factory=list)
    news: List[NewsArticle] = field(default_factory=list)


@dataclass
class Holding:
    symbol: str = ""
    quantity: int = 0
    avg_cost: float = 0.0
    current_price: float = 0.0
    market_value: float = 0.0
    pnl: float = 0.0
    pnl_percent: float = 0.0


@dataclass
class Portfolio:
    holdings: List[Holding] = field(default_factory=list)
    nav: float = 0.0
    currency: str = ""


@dataclass
class SectorContext:
    sector_code: str = ""
    sector_name: str = ""
    trend: str = ""
    sector_change_percent: float = 0.0
    stock_relative_performance: float = 0.0


@dataclass
class Observation:
    id: str = ""
    pattern_type: str = ""
    symbol: str = ""
    description: str = ""
    confidence: float = 0.0
    observed_at: str = ""
    outcome: str = ""


@dataclass
class Citation:
    source: str = ""
    url: str = ""
    claim: str = ""
    data_point: str = ""


# ---------------------------------------------------------------------------
# AnalyzeStock
# ---------------------------------------------------------------------------


@dataclass
class AnalyzeStockRequest:
    symbol: str = ""
    market_data: Optional[MarketData] = None
    portfolio: Optional[Portfolio] = None
    sector_context: Optional[SectorContext] = None
    knowledge_history: List[Observation] = field(default_factory=list)


@dataclass
class TechnicalAnalysis:
    symbol: str = ""
    composite_signal: str = ""
    indicators: Dict[str, float] = field(default_factory=dict)
    support_levels: List[str] = field(default_factory=list)
    resistance_levels: List[str] = field(default_factory=list)
    patterns: List[str] = field(default_factory=list)
    smart_money_flow: str = ""
    ma_crossovers: List[str] = field(default_factory=list)


@dataclass
class NewsAnalysis:
    symbol: str = ""
    sentiment: str = ""
    confidence: float = 0.0
    catalysts: List[str] = field(default_factory=list)
    risk_factors: List[str] = field(default_factory=list)
    articles: List[NewsArticle] = field(default_factory=list)


@dataclass
class InvestmentRecommendation:
    symbol: str = ""
    action: str = ""
    target_price: float = 0.0
    upside_percent: float = 0.0
    confidence_score: int = 0
    risk_level: str = ""
    reasoning: str = ""
    technical_factors: List[str] = field(default_factory=list)
    news_factors: List[str] = field(default_factory=list)


@dataclass
class TradingStrategy:
    symbol: str = ""
    signal_direction: str = ""
    entry_price: float = 0.0
    stop_loss: float = 0.0
    take_profit: float = 0.0
    risk_reward_ratio: float = 0.0
    confidence_score: int = 0
    position_size_percent: float = 0.0
    reasoning: str = ""


@dataclass
class AnalyzeStockResponse:
    technical: Optional[TechnicalAnalysis] = None
    news: Optional[NewsAnalysis] = None
    recommendation: Optional[InvestmentRecommendation] = None
    strategy: Optional[TradingStrategy] = None
    citations: List[Citation] = field(default_factory=list)
    disclaimer: str = ""


# ---------------------------------------------------------------------------
# GenerateInvestmentIdeas
# ---------------------------------------------------------------------------


@dataclass
class IdeaRequest:
    user_id: str = ""
    portfolio: Optional[Portfolio] = None
    watchlist_symbols: List[str] = field(default_factory=list)
    max_ideas: int = 0


@dataclass
class InvestmentIdea:
    symbol: str = ""
    signal_direction: str = ""
    entry_price: float = 0.0
    target_price: float = 0.0
    confidence_score: int = 0
    reasoning: str = ""
    historical_accuracy: float = 0.0


@dataclass
class IdeaResponse:
    ideas: List[InvestmentIdea] = field(default_factory=list)
    generated_at: str = ""


# ---------------------------------------------------------------------------
# Chat
# ---------------------------------------------------------------------------


@dataclass
class ChatMessage:
    role: str = ""
    content: str = ""
    timestamp: str = ""


@dataclass
class ProactiveSuggestion:
    text: str = ""
    type: str = ""
    symbol: str = ""


@dataclass
class ChatRequest:
    user_id: str = ""
    message: str = ""
    history: List[ChatMessage] = field(default_factory=list)
    portfolio: Optional[Portfolio] = None


@dataclass
class ChatResponse:
    response: str = ""
    citations: List[Citation] = field(default_factory=list)
    suggestions: List[ProactiveSuggestion] = field(default_factory=list)
    disclaimer: str = ""


# ---------------------------------------------------------------------------
# GetHotTopics
# ---------------------------------------------------------------------------


@dataclass
class HotTopicsRequest:
    limit: int = 0
    market: str = ""


@dataclass
class HotTopic:
    symbol: str = ""
    topic: str = ""
    category: str = ""
    relevance_score: float = 0.0
    summary: str = ""


@dataclass
class HotTopicsResponse:
    topics: List[HotTopic] = field(default_factory=list)
    generated_at: str = ""
