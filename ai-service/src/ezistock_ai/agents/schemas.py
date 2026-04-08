"""Pydantic models for LLM structured output via `model.with_structured_output()`.

These schemas are used with LangChain's `with_structured_output` to get typed,
validated responses from LLMs instead of raw JSON parsing.

Requirements: 4.2, 4.8, 5.4, 6.3, 7.2, 8.3
"""

from __future__ import annotations

from pydantic import BaseModel, Field


# ---------------------------------------------------------------------------
# Technical Analyst LLM output
# ---------------------------------------------------------------------------


class TechnicalNarrative(BaseModel):
    """Structured LLM output for technical analysis narrative."""

    composite_signal: str = Field(
        description="One of: strongly_bullish, bullish, neutral, bearish, strongly_bearish",
    )
    summary: str = Field(
        description="2-3 sentence overview of the technical picture",
    )
    indicator_analysis: dict[str, str] = Field(
        default_factory=dict,
        description="Mapping of indicator name to brief interpretation",
    )
    support_resistance_analysis: str = Field(
        default="",
        description="Brief interpretation of key support/resistance levels",
    )
    pattern_analysis: str = Field(
        default="",
        description="Brief interpretation of detected candlestick patterns",
    )
    crossover_analysis: str = Field(
        default="",
        description="Brief interpretation of MA crossovers",
    )
    smart_money_analysis: str = Field(
        default="",
        description="Brief interpretation of money flow",
    )
    overbought_oversold: OverboughtOversoldStatus = Field(
        default_factory=lambda: OverboughtOversoldStatus(),
        description="Overbought/oversold classification",
    )
    divergences: list[str] = Field(
        default_factory=list,
        description="Detected divergences between indicators and price",
    )
    key_risks: list[str] = Field(
        default_factory=list,
        description="Risk factors from the technical picture",
    )
    key_opportunities: list[str] = Field(
        default_factory=list,
        description="Opportunity factors from the technical picture",
    )


class OverboughtOversoldStatus(BaseModel):
    status: str = Field(default="neutral", description="overbought, oversold, or neutral")
    details: str = Field(default="", description="Supporting indicator values")


# ---------------------------------------------------------------------------
# News Analyst LLM output
# ---------------------------------------------------------------------------


class NewsSentimentOutput(BaseModel):
    """Structured LLM output for news sentiment analysis."""

    sentiment: str = Field(description="positive, negative, or neutral")
    confidence: float = Field(ge=0, le=1, description="Confidence level 0-1")
    catalysts: list[str] = Field(default_factory=list, description="Positive drivers")
    risk_factors: list[str] = Field(default_factory=list, description="Negative drivers")
    summary: str = Field(default="", description="2-3 sentence news summary")


# ---------------------------------------------------------------------------
# Investment Advisor LLM output
# ---------------------------------------------------------------------------


class InvestmentAdvice(BaseModel):
    """Structured LLM output for investment recommendations."""

    action: str = Field(description="buy, sell, or hold")
    target_price: float = Field(ge=0, description="Target price in VND")
    upside_percent: float = Field(description="Upside/downside percentage")
    confidence_score: int = Field(ge=0, le=100, description="Confidence 0-100")
    risk_level: str = Field(description="low, medium, or high")
    reasoning: str = Field(description="Structured reasoning referencing technical and news factors")
    technical_factors: list[str] = Field(default_factory=list)
    news_factors: list[str] = Field(default_factory=list)


# ---------------------------------------------------------------------------
# Strategy Builder LLM output
# ---------------------------------------------------------------------------


class TradingStrategyOutput(BaseModel):
    """Structured LLM output for trading strategy."""

    signal_direction: str = Field(description="long or short")
    entry_price: float = Field(ge=0, description="Entry price in VND")
    stop_loss: float = Field(ge=0, description="Stop-loss price (ATR-based)")
    take_profit: float = Field(ge=0, description="Take-profit price")
    risk_reward_ratio: float = Field(ge=0, description="Risk/reward ratio")
    confidence_score: int = Field(ge=0, le=100)
    position_size_percent: float = Field(
        ge=0, le=100,
        description="Suggested position size as % of NAV",
    )
    reasoning: str = Field(description="Strategy reasoning")


class InvestmentSignalOutput(BaseModel):
    """Structured LLM output for long-term investment signal."""

    entry_price_low: float = Field(ge=0, description="Entry zone low")
    entry_price_high: float = Field(ge=0, description="Entry zone high")
    target_price: float = Field(ge=0)
    holding_period: str = Field(description="Suggested holding period e.g. '3-6 months'")
    fundamental_reasoning: str = Field(default="")
    key_metrics: dict[str, str] = Field(default_factory=dict)
