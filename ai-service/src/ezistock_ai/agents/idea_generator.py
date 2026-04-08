"""Proactive Investment Idea Generator.

Daily scan of VN stock universe after market close to generate
buy/sell ideas with entry/SL/TP, confidence, and reasoning.
Applies liquidity filter and deduplication.

Requirements: 14.1, 14.2, 14.5, 14.6, 39.4
"""

from __future__ import annotations

import logging
import time
from dataclasses import dataclass, field

import numpy as np
import pandas as pd

logger = logging.getLogger(__name__)

_DEDUP_WINDOW_HOURS = 48
_DEDUP_CONFIDENCE_DELTA = 10
_MAX_IDEAS_PER_SCAN = 10


@dataclass
class InvestmentIdea:
    """A proactive buy/sell recommendation."""

    symbol: str
    direction: str  # "buy" or "sell"
    entry_price: float = 0.0
    stop_loss: float = 0.0
    take_profit: float = 0.0
    confidence: int = 0
    reasoning: str = ""
    trigger: str = ""  # What triggered this idea
    created_at: float = 0.0


@dataclass
class IdeaScanResult:
    """Output of a daily idea scan."""

    ideas: list[InvestmentIdea] = field(default_factory=list)
    scanned_symbols: int = 0
    filtered_out: int = 0  # Removed by liquidity filter
    deduplicated: int = 0  # Removed by dedup


class IdeaGenerator:
    """Scans the stock universe for actionable ideas.

    Detects:
    - Volume spikes (>2x 20d average)
    - Price gaps (>3% gap up/down)
    - Unusual foreign flow (>2 std dev from 20d mean)
    - Technical breakouts (price crossing key MAs)
    """

    def __init__(self) -> None:
        self._recent_ideas: list[InvestmentIdea] = []

    def scan(
        self,
        signal_space: pd.DataFrame,
        ohlcv: pd.DataFrame,
        liquidity_tiers: dict[str, int] | None = None,
    ) -> IdeaScanResult:
        """Run daily scan for investment ideas.

        Args:
            signal_space: Latest signal space from DataLayer.
            ohlcv: Recent OHLCV data with columns [date, symbol, open, high, low, close, volume].
            liquidity_tiers: Dict of symbol → tier (1/2/3). Tier 3 excluded.
        """
        result = IdeaScanResult()
        tiers = liquidity_tiers or {}

        if ohlcv.empty or signal_space.empty:
            return result

        symbols = ohlcv["symbol"].unique()
        result.scanned_symbols = len(symbols)
        raw_ideas: list[InvestmentIdea] = []

        for symbol in symbols:
            # Exclude Tier 3 (Req 39.4)
            if tiers.get(symbol, 1) >= 3:
                result.filtered_out += 1
                continue

            sym_data = ohlcv[ohlcv["symbol"] == symbol].sort_values("date")
            if len(sym_data) < 20:
                continue

            ideas = self._detect_outliers(symbol, sym_data, signal_space)
            raw_ideas.extend(ideas)

        # Sort by confidence descending
        raw_ideas.sort(key=lambda x: x.confidence, reverse=True)

        # Deduplication (Req 14.6): skip if same symbol+direction within 48h
        # unless confidence is 10+ points higher
        deduped = self._deduplicate(raw_ideas)
        result.deduplicated = len(raw_ideas) - len(deduped)

        result.ideas = deduped[:_MAX_IDEAS_PER_SCAN]

        # Track for future dedup
        self._recent_ideas.extend(result.ideas)
        # Prune old ideas
        cutoff = time.time() - _DEDUP_WINDOW_HOURS * 3600
        self._recent_ideas = [i for i in self._recent_ideas if i.created_at > cutoff]

        logger.info(
            "Idea scan: %d symbols, %d raw ideas, %d after dedup, %d filtered",
            result.scanned_symbols, len(raw_ideas), len(result.ideas), result.filtered_out,
        )
        return result

    def _detect_outliers(
        self,
        symbol: str,
        sym_data: pd.DataFrame,
        signal_space: pd.DataFrame,
    ) -> list[InvestmentIdea]:
        """Detect outlier conditions for a single symbol."""
        ideas: list[InvestmentIdea] = []
        latest = sym_data.iloc[-1]
        close = latest["close"]
        volume = latest["volume"]

        # Volume spike: >2x 20d average
        avg_vol_20d = sym_data["volume"].tail(20).mean()
        if avg_vol_20d > 0 and volume > 2 * avg_vol_20d:
            direction = "buy" if close > sym_data["close"].iloc[-2] else "sell"
            ideas.append(InvestmentIdea(
                symbol=symbol,
                direction=direction,
                entry_price=close,
                confidence=55,
                reasoning=f"Volume spike: {volume:,.0f} vs 20d avg {avg_vol_20d:,.0f} ({volume/avg_vol_20d:.1f}x)",
                trigger="volume_spike",
                created_at=time.time(),
            ))

        # Price gap: >3% gap from previous close
        prev_close = sym_data["close"].iloc[-2]
        gap_pct = (latest["open"] - prev_close) / prev_close if prev_close > 0 else 0
        if abs(gap_pct) > 0.03:
            direction = "buy" if gap_pct > 0 else "sell"
            ideas.append(InvestmentIdea(
                symbol=symbol,
                direction=direction,
                entry_price=close,
                confidence=50,
                reasoning=f"Price gap: {gap_pct:+.1%} from previous close {prev_close:,.0f}",
                trigger="price_gap",
                created_at=time.time(),
            ))

        # Foreign flow outlier (if available in signal space)
        try:
            dates = signal_space.index.get_level_values("date").unique()
            if len(dates) > 0 and "flow_foreign_net" in signal_space.columns:
                latest_date = dates[-1]
                sym_signals = signal_space.xs(latest_date, level="date")
                if symbol in sym_signals.index:
                    foreign_z = sym_signals.loc[symbol, "flow_foreign_net"]
                    if not np.isnan(foreign_z) and abs(foreign_z) > 2.0:
                        direction = "buy" if foreign_z > 0 else "sell"
                        ideas.append(InvestmentIdea(
                            symbol=symbol,
                            direction=direction,
                            entry_price=close,
                            confidence=60,
                            reasoning=f"Unusual foreign flow: z-score={foreign_z:.1f}",
                            trigger="foreign_flow",
                            created_at=time.time(),
                        ))
        except (KeyError, IndexError):
            pass

        # Set ATR-based SL/TP for all ideas
        atr = self._compute_atr(sym_data)
        for idea in ideas:
            if idea.direction == "buy":
                idea.stop_loss = round(close - 1.5 * atr, 2)
                idea.take_profit = round(close + 3.0 * atr, 2)
            else:
                idea.stop_loss = round(close + 1.5 * atr, 2)
                idea.take_profit = round(close - 3.0 * atr, 2)

        return ideas

    def _deduplicate(self, ideas: list[InvestmentIdea]) -> list[InvestmentIdea]:
        """Remove duplicate ideas within the dedup window (Req 14.6)."""
        cutoff = time.time() - _DEDUP_WINDOW_HOURS * 3600
        recent_keys = {
            (i.symbol, i.direction): i.confidence
            for i in self._recent_ideas
            if i.created_at > cutoff
        }

        result = []
        seen: set[tuple[str, str]] = set()
        for idea in ideas:
            key = (idea.symbol, idea.direction)
            if key in seen:
                continue

            prev_conf = recent_keys.get(key)
            if prev_conf is not None and idea.confidence < prev_conf + _DEDUP_CONFIDENCE_DELTA:
                continue  # Skip unless confidence is significantly higher

            seen.add(key)
            result.append(idea)

        return result

    @staticmethod
    def _compute_atr(df: pd.DataFrame, period: int = 14) -> float:
        """Compute ATR from OHLCV DataFrame."""
        if len(df) < period + 1:
            return 0.0
        high = df["high"].values
        low = df["low"].values
        close = df["close"].values
        tr = np.maximum(
            high[1:] - low[1:],
            np.maximum(
                np.abs(high[1:] - close[:-1]),
                np.abs(low[1:] - close[:-1]),
            ),
        )
        return float(np.mean(tr[-period:]))
