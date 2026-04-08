"""Investment Advisor Agent — synthesizes technical + news into recommendations.

Combines outputs from Technical Analyst and News Analyst, incorporates
portfolio context, sector context, and knowledge base history to produce
structured investment recommendations with confidence scores.

Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field

from ezistock_ai.agents.base import AgentContext, BaseAgent
from ezistock_ai.agents.news_analyst import NewsResult
from ezistock_ai.agents.schemas import InvestmentAdvice
from ezistock_ai.agents.technical_analyst import TechnicalResult
from ezistock_ai.generated.proto.agent_pb2 import InvestmentRecommendation
from ezistock_ai.llm.router import LLMRouter, TaskType

logger = logging.getLogger(__name__)


@dataclass
class AdvisorResult:
    """Internal result from the investment advisor."""

    symbol: str
    action: str = "hold"
    target_price: float = 0.0
    upside_percent: float = 0.0
    confidence_score: int = 0
    risk_level: str = "medium"
    reasoning: str = ""
    technical_factors: list[str] = field(default_factory=list)
    news_factors: list[str] = field(default_factory=list)


class AsyncInvestmentAdvisor(BaseAgent[AdvisorResult]):
    """Synthesizes technical and news analysis into investment recommendations.

    Requires Phase 1 outputs (technical + news) in ctx.extra.
    """

    def __init__(
        self,
        llm_router: LLMRouter | None = None,
        timeout: float = 30.0,
    ) -> None:
        super().__init__(name="investment_advisor", timeout=timeout)
        self._llm_router = llm_router

    async def _run(self, ctx: AgentContext) -> AdvisorResult:
        symbol = ctx.symbol
        technical: TechnicalResult | None = ctx.extra.get("technical_result")
        news: NewsResult | None = ctx.extra.get("news_result")

        if technical is None and news is None:
            logger.warning("No Phase 1 data for %s, returning neutral hold", symbol)
            return AdvisorResult(
                symbol=symbol,
                reasoning="Insufficient data — both technical and news analysis unavailable.",
            )

        # Build context strings for the prompt
        current_price = self._get_current_price(ctx)
        indicators_summary = self._summarize_indicators(technical)
        portfolio_context = self._build_portfolio_context(ctx)
        sector_context = self._build_sector_context(ctx)

        # Add citations
        if technical:
            self.citation_collector.add(
                source="Technical Analysis",
                claim=f"Composite signal: {technical.composite_signal.value}",
                data_point=f"Bullish: {technical.bullish_count}, Bearish: {technical.bearish_count}",
            )
        if news and news.sentiment != "neutral":
            self.citation_collector.add(
                source="News Analysis",
                claim=f"News sentiment: {news.sentiment}",
                data_point=f"Confidence: {news.confidence:.0%}",
            )

        if self._llm_router is not None:
            return await self._advise_with_llm(
                symbol, technical, news, current_price,
                indicators_summary, portfolio_context, sector_context, ctx,
            )

        # Fallback: rule-based recommendation without LLM
        return self._rule_based_recommendation(symbol, technical, news, current_price)

    async def _advise_with_llm(
        self,
        symbol: str,
        technical: TechnicalResult | None,
        news: NewsResult | None,
        current_price: float,
        indicators_summary: str,
        portfolio_context: str,
        sector_context: str,
        ctx: AgentContext,
    ) -> AdvisorResult:
        from ezistock_ai.agents.prompts.advisor import ANALYSIS_PROMPT, SYSTEM_PROMPT

        model = self._llm_router.get_model(TaskType.ANALYSIS)
        structured_model = model.with_structured_output(InvestmentAdvice)

        messages = [
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": ANALYSIS_PROMPT.format(
                symbol=symbol,
                composite_signal=technical.composite_signal.value if technical else "unavailable",
                indicators_summary=indicators_summary,
                support_levels=", ".join(f"{s.level:.0f}" for s in (technical.support_levels[:3] if technical else [])) or "N/A",
                resistance_levels=", ".join(f"{r.level:.0f}" for r in (technical.resistance_levels[:3] if technical else [])) or "N/A",
                patterns=", ".join(f"{p.name} ({p.direction})" for p in (technical.patterns if technical else [])) or "None detected",
                smart_money_flow=technical.smart_money_flow.value if technical else "unavailable",
                news_sentiment=news.sentiment if news else "unavailable",
                news_confidence=news.confidence if news else 0.0,
                catalysts=", ".join(news.catalysts[:5]) if news and news.catalysts else "None identified",
                risk_factors=", ".join(news.risk_factors[:5]) if news and news.risk_factors else "None identified",
                portfolio_context=portfolio_context,
                sector_context=sector_context,
                accuracy_context=ctx.accuracy_context or "No historical accuracy data available",
                current_price=f"{current_price:,.0f}" if current_price > 0 else "Unknown",
            )},
        ]

        output: InvestmentAdvice = await structured_model.ainvoke(messages)

        return AdvisorResult(
            symbol=symbol,
            action=output.action,
            target_price=output.target_price,
            upside_percent=output.upside_percent,
            confidence_score=output.confidence_score,
            risk_level=output.risk_level,
            reasoning=output.reasoning,
            technical_factors=output.technical_factors,
            news_factors=output.news_factors,
        )

    @staticmethod
    def _rule_based_recommendation(
        symbol: str,
        technical: TechnicalResult | None,
        news: NewsResult | None,
        current_price: float,
    ) -> AdvisorResult:
        """Simple rule-based fallback when no LLM is available."""
        action = "hold"
        confidence = 30
        risk = "medium"
        reasons = []

        if technical:
            sig = technical.composite_signal.value
            if "bullish" in sig:
                action = "buy"
                confidence += 20
                reasons.append(f"Technical: {sig}")
            elif "bearish" in sig:
                action = "sell"
                confidence += 20
                reasons.append(f"Technical: {sig}")

        if news and news.sentiment != "neutral":
            if news.sentiment == "positive" and action != "sell":
                action = "buy"
                confidence += int(news.confidence * 15)
                reasons.append(f"News: {news.sentiment}")
            elif news.sentiment == "negative" and action != "buy":
                action = "sell"
                confidence += int(news.confidence * 15)
                reasons.append(f"News: {news.sentiment}")

        return AdvisorResult(
            symbol=symbol,
            action=action,
            confidence_score=min(confidence, 100),
            risk_level=risk,
            reasoning="; ".join(reasons) or "Insufficient data for strong conviction",
        )

    @staticmethod
    def _get_current_price(ctx: AgentContext) -> float:
        if ctx.market_data and ctx.market_data.ohlcv:
            return ctx.market_data.ohlcv[-1].close
        return 0.0

    @staticmethod
    def _summarize_indicators(technical: TechnicalResult | None) -> str:
        if not technical or not technical.indicators:
            return "No technical indicators available"
        top = [f"{ind.name}: {ind.value:.2f} ({ind.signal})" for ind in technical.indicators[:8] if ind.value == ind.value]
        return "; ".join(top)

    @staticmethod
    def _build_portfolio_context(ctx: AgentContext) -> str:
        if ctx.market_data and hasattr(ctx, "extra"):
            portfolio = ctx.extra.get("portfolio")
            if portfolio:
                return f"User holds portfolio with NAV={portfolio.nav:,.0f} VND"
        return "No portfolio data available"

    @staticmethod
    def _build_sector_context(ctx: AgentContext) -> str:
        if ctx.market_data and hasattr(ctx, "extra"):
            sector = ctx.extra.get("sector_context")
            if sector:
                return f"Sector: {sector.sector_name}, Trend: {sector.trend}, Change: {sector.sector_change_percent:.1f}%"
        return "No sector context available"

    @staticmethod
    def to_protobuf(result: AdvisorResult) -> InvestmentRecommendation:
        return InvestmentRecommendation(
            symbol=result.symbol,
            action=result.action,
            target_price=result.target_price,
            upside_percent=result.upside_percent,
            confidence_score=result.confidence_score,
            risk_level=result.risk_level,
            reasoning=result.reasoning,
            technical_factors=result.technical_factors,
            news_factors=result.news_factors,
        )
