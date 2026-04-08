"""Agent Orchestrator — LangGraph workflow for multi-agent pipeline.

Implements parallel agent execution with 30s per-agent timeout,
partial failure handling, citation aggregation, and structured
degradation metadata.

Requirements: 4.1, 4.2, 4.6, 4.8
"""

from __future__ import annotations

import asyncio
import logging
from dataclasses import dataclass, field

from ezistock_ai.agents.base import (
    AgentContext,
    AgentResult,
    BaseAgent,
    DegradationInfo,
)
from ezistock_ai.agents.investment_advisor import AdvisorResult, AsyncInvestmentAdvisor
from ezistock_ai.agents.memory import AgentMemory
from ezistock_ai.agents.news_analyst import AsyncNewsAnalyst, NewsResult
from ezistock_ai.agents.strategy_builder import AsyncStrategyBuilder, StrategyResult
from ezistock_ai.agents.technical_analyst import TechnicalAnalystAgent, TechnicalResult
from ezistock_ai.generated.proto.agent_pb2 import (
    AnalyzeStockRequest,
    AnalyzeStockResponse,
    Citation,
    TechnicalAnalysis,
)
from ezistock_ai.llm.cache import LLMCache
from ezistock_ai.llm.router import LLMRouter

logger = logging.getLogger(__name__)

# Confidence penalty per missing agent
_CONFIDENCE_PENALTY = {
    "technical_analyst": -20,
    "news_analyst": -15,
    "investment_advisor": -25,
    "strategy_builder": -10,
}

# Cache TTL for full analysis results (1 hour, per Req 48.4)
_ANALYSIS_CACHE_TTL = 3600.0


# ---------------------------------------------------------------------------
# Async wrapper for the TechnicalAnalystAgent (sync → async)
# ---------------------------------------------------------------------------


class AsyncTechnicalAnalyst(BaseAgent[TechnicalResult]):
    """Wraps the sync TechnicalAnalystAgent in the BaseAgent async interface."""

    def __init__(self, timeout: float = 30.0) -> None:
        super().__init__(name="technical_analyst", timeout=timeout)
        self._engine = TechnicalAnalystAgent()

    async def _run(self, ctx: AgentContext) -> TechnicalResult:
        # Run CPU-bound indicator computation in a thread pool
        loop = asyncio.get_running_loop()
        result = await loop.run_in_executor(
            None,
            self._engine.analyze,
            ctx.market_data,
            ctx.extra.get("foreign_net_volume", 0.0),
            ctx.extra.get("institutional_net_volume", 0.0),
        )

        # Populate citations from computed indicators
        for ind in result.indicators:
            if ind.detail:
                self.citation_collector.add(
                    source=f"OHLCV data for {ctx.symbol}",
                    claim=ind.detail,
                    data_point=f"{ind.name}={ind.value:.4f}" if not _is_nan(ind.value) else ind.name,
                )

        return result


# ---------------------------------------------------------------------------
# Orchestrator
# ---------------------------------------------------------------------------


@dataclass
class OrchestratorResult:
    """Full output of the multi-agent pipeline."""

    technical: TechnicalResult | None = None
    news: NewsResult | None = None
    recommendation: AdvisorResult | None = None
    strategy: StrategyResult | None = None
    citations: list[Citation] = field(default_factory=list)
    degradation: DegradationInfo = field(default_factory=DegradationInfo)
    disclaimer: str = "This is AI-generated analysis, not financial advice."


class AgentOrchestrator:
    """Coordinates the multi-agent pipeline with parallel execution.

    Phase 1 (parallel): Technical Analyst + News Analyst
    Phase 2 (sequential): Investment Advisor (needs Phase 1 outputs)
    Phase 3 (sequential): Strategy Builder (needs Phase 2 output)

    If any agent fails or times out, the pipeline continues with available
    results and records degradation metadata.
    """

    def __init__(self, llm_router: LLMRouter | None = None) -> None:
        self._technical = AsyncTechnicalAnalyst()
        self._news = AsyncNewsAnalyst(llm_router=llm_router)
        self._advisor = AsyncInvestmentAdvisor(llm_router=llm_router)
        self._strategy = AsyncStrategyBuilder(llm_router=llm_router)
        self._memory = AgentMemory()
        self._cache = LLMCache(default_ttl=_ANALYSIS_CACHE_TTL)

    async def analyze(self, request: AnalyzeStockRequest) -> OrchestratorResult:
        """Run the full multi-agent pipeline for a single stock."""

        # --- Cache check: return cached result if OHLCV data hasn't changed ---
        cache_inputs = self._build_cache_key(request)
        cached = self._cache.get("orchestrator", cache_inputs)
        if cached is not None:
            logger.info("Cache hit for %s — returning cached analysis", request.symbol)
            return cached

        ctx = AgentContext(
            symbol=request.symbol,
            market_data=request.market_data,
            extra={},
        )

        result = OrchestratorResult()

        # --- Pre-flight: Load memory context (best-effort) ---
        try:
            memory_ctx = await self._memory.load_context(request.symbol)
            ctx.accuracy_context = memory_ctx.to_prompt_context()
            ctx.historical_accuracy = (
                next(iter(memory_ctx.agent_accuracy.values()), None)
                if memory_ctx.agent_accuracy else None
            )
        except Exception as exc:
            logger.debug("Memory load failed (non-fatal): %s", exc)

        # --- Phase 1: Parallel execution (Technical + News) ---
        phase1_agents: list[BaseAgent] = [self._technical, self._news]

        phase1_results = await self._run_parallel(phase1_agents, ctx)

        for agent_result in phase1_results:
            if agent_result.agent_name == "technical_analyst":
                if agent_result.success and agent_result.output is not None:
                    result.technical = agent_result.output
                else:
                    result.degradation.missing_agents.append("technical_analyst")
                    result.degradation.reasons["technical_analyst"] = agent_result.error or "unknown"
                result.citations.extend(agent_result.citations)

            elif agent_result.agent_name == "news_analyst":
                if agent_result.success and agent_result.output is not None:
                    result.news = agent_result.output
                else:
                    result.degradation.missing_agents.append("news_analyst")
                    result.degradation.reasons["news_analyst"] = agent_result.error or "unknown"
                result.citations.extend(agent_result.citations)

        # --- Phase 2: Investment Advisor (needs Phase 1) ---
        if result.technical is not None or result.news is not None:
            ctx.extra["technical_result"] = result.technical
            ctx.extra["news_result"] = result.news
            if request.portfolio:
                ctx.extra["portfolio"] = request.portfolio
            if request.sector_context:
                ctx.extra["sector_context"] = request.sector_context

            advisor_result = await self._advisor.run(ctx)
            if advisor_result.success and advisor_result.output is not None:
                result.recommendation = advisor_result.output
            else:
                result.degradation.missing_agents.append("investment_advisor")
                result.degradation.reasons["investment_advisor"] = advisor_result.error or "unknown"
            result.citations.extend(advisor_result.citations)
        else:
            result.degradation.missing_agents.append("investment_advisor")
            result.degradation.reasons["investment_advisor"] = "No Phase 1 data available"

        # --- Phase 3: Strategy Builder (needs Phase 2) ---
        if result.recommendation is not None:
            ctx.extra["advisor_result"] = result.recommendation

            strategy_result = await self._strategy.run(ctx)
            if strategy_result.success and strategy_result.output is not None:
                result.strategy = strategy_result.output
            else:
                result.degradation.missing_agents.append("strategy_builder")
                result.degradation.reasons["strategy_builder"] = strategy_result.error or "unknown"
            result.citations.extend(strategy_result.citations)
        else:
            result.degradation.missing_agents.append("strategy_builder")
            result.degradation.reasons["strategy_builder"] = "No advisor recommendation available"

        # Compute confidence adjustment from missing agents
        for missing in result.degradation.missing_agents:
            result.degradation.confidence_adjustment += _CONFIDENCE_PENALTY.get(missing, -10)

        logger.info(
            "Orchestrator completed for %s: tech=%s news=%s advisor=%s strategy=%s missing=%s",
            request.symbol,
            "ok" if result.technical else "fail",
            "ok" if result.news else "fail",
            "ok" if result.recommendation else "fail",
            "ok" if result.strategy else "fail",
            result.degradation.missing_agents or "none",
        )

        # --- Post-flight: Persist recommendation to memory (best-effort) ---
        if result.recommendation is not None:
            try:
                await self._memory.store_recommendation(
                    symbol=request.symbol,
                    action=result.recommendation.action,
                    target_price=result.recommendation.target_price,
                    confidence_score=result.recommendation.confidence_score,
                    risk_level=result.recommendation.risk_level,
                    reasoning=result.recommendation.reasoning,
                    agent_outputs={
                        "composite_signal": result.technical.composite_signal.value if result.technical else None,
                        "news_sentiment": result.news.sentiment if result.news else None,
                    },
                )
            except Exception as exc:
                logger.debug("Memory store failed (non-fatal): %s", exc)

        # Store notable technical patterns as observations
        if result.technical and result.technical.patterns:
            for pattern in result.technical.patterns[:3]:
                try:
                    await self._memory.store_observation(
                        symbol=request.symbol,
                        pattern_type=pattern.name,
                        description=f"{pattern.name} ({pattern.direction}) detected",
                        confidence=0.7,
                    )
                except Exception:
                    pass

        # --- Cache the result for subsequent identical requests ---
        self._cache.put("orchestrator", cache_inputs, result)

        return result

    def to_protobuf(self, result: OrchestratorResult) -> AnalyzeStockResponse:
        """Convert OrchestratorResult to protobuf AnalyzeStockResponse."""
        technical_pb = None
        if result.technical is not None:
            technical_pb = self._technical._engine.to_protobuf(result.technical)

        news_pb = None
        if result.news is not None:
            news_pb = AsyncNewsAnalyst.to_protobuf(result.news)

        recommendation_pb = None
        if result.recommendation is not None:
            recommendation_pb = AsyncInvestmentAdvisor.to_protobuf(result.recommendation)

        strategy_pb = None
        if result.strategy is not None:
            strategy_pb = AsyncStrategyBuilder.to_protobuf(result.strategy)

        return AnalyzeStockResponse(
            technical=technical_pb,
            news=news_pb,
            recommendation=recommendation_pb,
            strategy=strategy_pb,
            citations=result.citations,
            disclaimer=result.disclaimer,
        )

    # ------------------------------------------------------------------
    # Private helpers
    # ------------------------------------------------------------------

    @staticmethod
    async def _run_parallel(
        agents: list[BaseAgent],
        ctx: AgentContext,
    ) -> list[AgentResult]:
        """Run multiple agents concurrently via asyncio.gather.

        Each agent has its own internal timeout (default 30s).
        gather(return_exceptions=False) is NOT used — each agent's
        run() method already catches all exceptions internally.
        """
        if not agents:
            return []

        tasks = [agent.run(ctx) for agent in agents]
        results = await asyncio.gather(*tasks)
        return list(results)

    @staticmethod
    def _build_cache_key(request: AnalyzeStockRequest) -> dict:
        """Build a deterministic cache key from the request.

        Keyed by symbol + last OHLCV bar date + close price, so the cache
        invalidates naturally when new market data arrives.
        """
        last_bar_date = ""
        last_close = 0.0
        bar_count = 0
        if request.market_data and request.market_data.ohlcv:
            bars = request.market_data.ohlcv
            bar_count = len(bars)
            last_bar_date = bars[-1].date
            last_close = bars[-1].close

        return {
            "symbol": request.symbol,
            "last_bar_date": last_bar_date,
            "last_close": last_close,
            "bar_count": bar_count,
        }

    @property
    def cache_stats(self) -> dict:
        """Expose cache hit/miss stats for monitoring."""
        return self._cache.stats


def _is_nan(val: float) -> bool:
    """Check for NaN without importing numpy."""
    return val != val
