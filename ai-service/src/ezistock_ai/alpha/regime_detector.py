"""Alpha Mining Engine — Regime Detector.

Classifies the Vietnamese market into regimes (bull, bear, sideways,
risk-on, risk-off) using VN-Index trend, market breadth, foreign flow,
and volatility levels.

Requirements: 10.4
"""

from __future__ import annotations

import logging
from dataclasses import dataclass
from enum import Enum

import numpy as np
import pandas as pd

logger = logging.getLogger(__name__)


class MarketRegime(str, Enum):
    BULL = "bull"
    BEAR = "bear"
    SIDEWAYS = "sideways"
    RISK_ON = "risk_on"
    RISK_OFF = "risk_off"


@dataclass
class RegimeSignals:
    """Raw signals used for regime classification."""

    vn_index_sma50_above_sma200: bool = False
    vn_index_20d_return: float = 0.0
    market_breadth: float = 0.5  # % of stocks above their 50d SMA
    foreign_flow_20d: float = 0.0  # Net foreign flow over 20 days
    volatility_percentile: float = 0.5  # Current vol vs 1-year range


@dataclass
class RegimeResult:
    """Output of regime detection."""

    regime: MarketRegime
    confidence: float  # 0-1
    signals: RegimeSignals
    description: str


class RegimeDetector:
    """Classifies current VN market regime from index and breadth data."""

    def detect(
        self,
        vn_index: pd.DataFrame,
        breadth_ratio: float = 0.5,
        foreign_flow_20d: float = 0.0,
    ) -> RegimeResult:
        """Classify the current market regime.

        Args:
            vn_index: DataFrame with columns [date, close] for VN-Index.
            breadth_ratio: Fraction of stocks above their 50d SMA (0-1).
            foreign_flow_20d: Net foreign buy-sell volume over last 20 days.

        Returns:
            RegimeResult with classified regime and confidence.
        """
        if vn_index.empty or len(vn_index) < 200:
            return RegimeResult(
                regime=MarketRegime.SIDEWAYS,
                confidence=0.3,
                signals=RegimeSignals(),
                description="Insufficient data for regime detection",
            )

        close = vn_index["close"].values
        n = len(close)

        # Compute signals
        sma50 = np.mean(close[-50:])
        sma200 = np.mean(close[-200:])
        sma50_above_200 = sma50 > sma200

        ret_20d = (close[-1] / close[-20] - 1) if n >= 20 else 0.0

        # Volatility: 20d realized vol vs 1-year range
        if n >= 252:
            returns = np.diff(close[-252:]) / close[-253:-1]
            vol_1y = np.std(returns)
            vol_20d = np.std(returns[-20:]) if len(returns) >= 20 else vol_1y
            vol_percentile = 0.0
            if vol_1y > 0:
                # Where does current vol sit in the 1-year range
                rolling_vols = pd.Series(returns).rolling(20).std().dropna().values
                vol_percentile = float(np.searchsorted(np.sort(rolling_vols), vol_20d) / len(rolling_vols))
        else:
            vol_percentile = 0.5

        signals = RegimeSignals(
            vn_index_sma50_above_sma200=sma50_above_200,
            vn_index_20d_return=ret_20d,
            market_breadth=breadth_ratio,
            foreign_flow_20d=foreign_flow_20d,
            volatility_percentile=vol_percentile,
        )

        # Classification logic
        regime, confidence, desc = self._classify(signals)

        return RegimeResult(
            regime=regime,
            confidence=confidence,
            signals=signals,
            description=desc,
        )

    @staticmethod
    def _classify(s: RegimeSignals) -> tuple[MarketRegime, float, str]:
        """Rule-based regime classification with confidence scoring."""
        scores: dict[MarketRegime, float] = {r: 0.0 for r in MarketRegime}

        # Bull signals
        if s.vn_index_sma50_above_sma200:
            scores[MarketRegime.BULL] += 0.3
        if s.vn_index_20d_return > 0.03:
            scores[MarketRegime.BULL] += 0.2
        if s.market_breadth > 0.6:
            scores[MarketRegime.BULL] += 0.2
        if s.foreign_flow_20d > 0:
            scores[MarketRegime.BULL] += 0.1
            scores[MarketRegime.RISK_ON] += 0.15

        # Bear signals
        if not s.vn_index_sma50_above_sma200:
            scores[MarketRegime.BEAR] += 0.3
        if s.vn_index_20d_return < -0.03:
            scores[MarketRegime.BEAR] += 0.2
        if s.market_breadth < 0.4:
            scores[MarketRegime.BEAR] += 0.2
        if s.foreign_flow_20d < 0:
            scores[MarketRegime.BEAR] += 0.1
            scores[MarketRegime.RISK_OFF] += 0.15

        # Sideways
        if abs(s.vn_index_20d_return) < 0.02:
            scores[MarketRegime.SIDEWAYS] += 0.3
        if 0.4 <= s.market_breadth <= 0.6:
            scores[MarketRegime.SIDEWAYS] += 0.2

        # Risk-on / Risk-off from volatility
        if s.volatility_percentile < 0.3:
            scores[MarketRegime.RISK_ON] += 0.2
        elif s.volatility_percentile > 0.7:
            scores[MarketRegime.RISK_OFF] += 0.25

        # Pick highest scoring regime
        regime = max(scores, key=scores.get)  # type: ignore
        confidence = min(scores[regime], 1.0)

        descriptions = {
            MarketRegime.BULL: f"Bullish: SMA50>SMA200={s.vn_index_sma50_above_sma200}, 20d return={s.vn_index_20d_return:.1%}, breadth={s.market_breadth:.0%}",
            MarketRegime.BEAR: f"Bearish: SMA50>SMA200={s.vn_index_sma50_above_sma200}, 20d return={s.vn_index_20d_return:.1%}, breadth={s.market_breadth:.0%}",
            MarketRegime.SIDEWAYS: f"Sideways: 20d return={s.vn_index_20d_return:.1%}, breadth={s.market_breadth:.0%}",
            MarketRegime.RISK_ON: f"Risk-on: Low volatility (percentile={s.volatility_percentile:.0%}), positive foreign flow",
            MarketRegime.RISK_OFF: f"Risk-off: High volatility (percentile={s.volatility_percentile:.0%}), negative foreign flow",
        }

        return regime, confidence, descriptions[regime]
