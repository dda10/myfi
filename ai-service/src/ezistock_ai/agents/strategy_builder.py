"""Strategy Builder Agent — develops trading strategies with entry/exit timing.

Produces Trading_Signal and Investment_Signal outputs based on multi-agent
analysis, ATR-based stop-loss, market regime, and liquidity tier adjustments.

Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 39.7, 39.8
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field

from ezistock_ai.agents.base import AgentContext, BaseAgent
from ezistock_ai.agents.investment_advisor import AdvisorResult
from ezistock_ai.agents.schemas import TradingStrategyOutput
from ezistock_ai.agents.technical_analyst import TechnicalResult
from ezistock_ai.generated.proto.agent_pb2 import TradingStrategy
from ezistock_ai.llm.router import LLMRouter, TaskType

logger = logging.getLogger(__name__)

# Default risk/reward targets
_DEFAULT_RR_RATIO = 2.0
_ATR_STOP_MULTIPLIER = 1.5
_ATR_TARGET_MULTIPLIER = 3.0
_TIER2_POSITION_REDUCTION = 0.5


@dataclass
class StrategyResult:
    """Internal result from the strategy builder."""

    symbol: str
    signal_direction: str = "long"
    entry_price: float = 0.0
    stop_loss: float = 0.0
    take_profit: float = 0.0
    risk_reward_ratio: float = 0.0
    confidence_score: int = 0
    position_size_percent: float = 0.0
    reasoning: str = ""


class AsyncStrategyBuilder(BaseAgent[StrategyResult]):
    """Builds trading strategies from advisor recommendations + technical data.

    Requires Phase 2 outputs (advisor result + technical) in ctx.extra.
    """

    def __init__(
        self,
        llm_router: LLMRouter | None = None,
        timeout: float = 30.0,
    ) -> None:
        super().__init__(name="strategy_builder", timeout=timeout)
        self._llm_router = llm_router

    async def _run(self, ctx: AgentContext) -> StrategyResult:
        symbol = ctx.symbol
        advisor: AdvisorResult | None = ctx.extra.get("advisor_result")
        technical: TechnicalResult | None = ctx.extra.get("technical_result")
        liquidity_tier: int = ctx.extra.get("liquidity_tier", 1)
        regime: str = ctx.extra.get("market_regime", "unknown")

        # Reject Tier 3 stocks (Req 39.8)
        if liquidity_tier >= 3:
            logger.info("Rejecting %s — Tier 3 (illiquid)", symbol)
            return StrategyResult(
                symbol=symbol,
                reasoning=f"Excluded: {symbol} is Tier 3 (illiquid). No strategy generated.",
            )

        current_price = self._get_current_price(ctx)
        if current_price <= 0:
            return StrategyResult(symbol=symbol, reasoning="No price data available")

        # Get ATR for stop-loss calculation
        atr_val = self._get_atr(technical)

        if self._llm_router is not None and advisor is not None:
            return await self._build_with_llm(
                symbol, advisor, technical, current_price,
                atr_val, liquidity_tier, regime, ctx,
            )

        # Fallback: rule-based strategy
        return self._rule_based_strategy(
            symbol, advisor, current_price, atr_val, liquidity_tier, regime,
        )

    async def _build_with_llm(
        self,
        symbol: str,
        advisor: AdvisorResult,
        technical: TechnicalResult | None,
        current_price: float,
        atr_val: float,
        liquidity_tier: int,
        regime: str,
        ctx: AgentContext,
    ) -> StrategyResult:
        from ezistock_ai.agents.prompts.strategy import STRATEGY_PROMPT, SYSTEM_PROMPT

        model = self._llm_router.get_model(TaskType.ANALYSIS)
        structured_model = model.with_structured_output(TradingStrategyOutput)

        tier_label = {1: "Tier 1 (highly liquid)", 2: "Tier 2 (moderate — 50% position reduction)", 3: "Tier 3 (illiquid — excluded)"}

        messages = [
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": STRATEGY_PROMPT.format(
                symbol=symbol,
                action=advisor.action,
                target_price=advisor.target_price or current_price,
                confidence_score=advisor.confidence_score,
                risk_level=advisor.risk_level,
                current_price=current_price,
                atr=atr_val,
                support_levels=", ".join(f"{s.level:.0f}" for s in (technical.support_levels[:3] if technical else [])) or "N/A",
                resistance_levels=", ".join(f"{r.level:.0f}" for r in (technical.resistance_levels[:3] if technical else [])) or "N/A",
                composite_signal=technical.composite_signal.value if technical else "unavailable",
                regime_info=f"Current regime: {regime}" if regime != "unknown" else "Regime data unavailable",
                liquidity_tier=tier_label.get(liquidity_tier, f"Tier {liquidity_tier}"),
            )},
        ]

        output: TradingStrategyOutput = await structured_model.ainvoke(messages)

        # Apply Tier 2 position reduction (Req 39.7)
        position_size = output.position_size_percent
        if liquidity_tier == 2:
            position_size *= _TIER2_POSITION_REDUCTION

        self.citation_collector.add(
            source="Strategy Builder",
            claim=f"Entry: {output.entry_price:,.0f}, SL: {output.stop_loss:,.0f}, TP: {output.take_profit:,.0f}",
            data_point=f"ATR={atr_val:.2f}, R:R={output.risk_reward_ratio:.1f}",
        )

        return StrategyResult(
            symbol=symbol,
            signal_direction=output.signal_direction,
            entry_price=output.entry_price,
            stop_loss=output.stop_loss,
            take_profit=output.take_profit,
            risk_reward_ratio=output.risk_reward_ratio,
            confidence_score=output.confidence_score,
            position_size_percent=position_size,
            reasoning=output.reasoning,
        )

    @staticmethod
    def _rule_based_strategy(
        symbol: str,
        advisor: AdvisorResult | None,
        current_price: float,
        atr_val: float,
        liquidity_tier: int,
        regime: str,
    ) -> StrategyResult:
        """ATR-based strategy without LLM."""
        action = advisor.action if advisor else "hold"
        confidence = advisor.confidence_score if advisor else 30

        if action == "hold" or atr_val <= 0:
            return StrategyResult(
                symbol=symbol,
                reasoning="Hold recommendation — no active trading strategy.",
            )

        direction = "long" if action == "buy" else "short"

        # ATR-based levels
        if direction == "long":
            stop_loss = current_price - _ATR_STOP_MULTIPLIER * atr_val
            take_profit = current_price + _ATR_TARGET_MULTIPLIER * atr_val
        else:
            stop_loss = current_price + _ATR_STOP_MULTIPLIER * atr_val
            take_profit = current_price - _ATR_TARGET_MULTIPLIER * atr_val

        risk = abs(current_price - stop_loss)
        reward = abs(take_profit - current_price)
        rr_ratio = reward / risk if risk > 0 else 0.0

        # Position sizing: base 5% of NAV, adjust for regime and tier
        base_position = 5.0
        if regime in ("bear", "risk-off"):
            base_position *= 0.6
        elif regime in ("bull", "risk-on"):
            base_position *= 1.2
        if liquidity_tier == 2:
            base_position *= _TIER2_POSITION_REDUCTION

        return StrategyResult(
            symbol=symbol,
            signal_direction=direction,
            entry_price=current_price,
            stop_loss=round(stop_loss, 2),
            take_profit=round(take_profit, 2),
            risk_reward_ratio=round(rr_ratio, 2),
            confidence_score=confidence,
            position_size_percent=round(base_position, 1),
            reasoning=f"ATR-based {direction} strategy. SL={_ATR_STOP_MULTIPLIER}×ATR, TP={_ATR_TARGET_MULTIPLIER}×ATR. Regime: {regime}.",
        )

    @staticmethod
    def _get_current_price(ctx: AgentContext) -> float:
        if ctx.market_data and ctx.market_data.ohlcv:
            return ctx.market_data.ohlcv[-1].close
        return 0.0

    @staticmethod
    def _get_atr(technical: TechnicalResult | None) -> float:
        if not technical:
            return 0.0
        for ind in technical.indicators:
            if ind.name == "ATR(14)" and ind.value == ind.value:  # NaN check
                return ind.value
        return 0.0

    @staticmethod
    def to_protobuf(result: StrategyResult) -> TradingStrategy:
        return TradingStrategy(
            symbol=result.symbol,
            signal_direction=result.signal_direction,
            entry_price=result.entry_price,
            stop_loss=result.stop_loss,
            take_profit=result.take_profit,
            risk_reward_ratio=result.risk_reward_ratio,
            confidence_score=result.confidence_score,
            position_size_percent=result.position_size_percent,
            reasoning=result.reasoning,
        )
