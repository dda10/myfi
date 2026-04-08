"""Alpha Mining Engine — Deployment Layer.

Strategy Ensemble with consensus voting, Alpha Decay Monitor with
20-trading-day monitoring window (VN market faster decay), and
continuous learning feedback to Model Layer.

Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field

import numpy as np
import pandas as pd

logger = logging.getLogger(__name__)

_DEFAULT_CONSENSUS_MIN = 2  # Min strategies that must agree (Req 12.2)
_DECAY_WINDOW = 20  # Trading days for VN market (Req 12.5)
_DECAY_IC_THRESHOLD = 0.01


@dataclass
class RankedStock:
    """A stock in the consensus ranking."""

    symbol: str
    composite_score: float  # Weighted average across strategies
    agreement_count: int    # How many strategies rank it in top N
    strategy_scores: dict[str, float] = field(default_factory=dict)


@dataclass
class EnsembleResult:
    """Output of the Strategy Ensemble (Req 12.1, 12.3)."""

    ranking: list[RankedStock] = field(default_factory=list)
    consensus_threshold: int = _DEFAULT_CONSENSUS_MIN
    total_strategies: int = 0
    decay_alerts: list[str] = field(default_factory=list)


@dataclass
class DecayAlert:
    """Alert when a signal's predictive power degrades."""

    signal_name: str
    ic_current: float
    ic_peak: float
    decay_rate: float
    action: str  # "rebalance", "replace", "monitor"


class StrategyEnsemble:
    """Combines multiple strategy predictions via consensus voting (Req 12.1).

    A stock only enters the top ranking if at least `min_agreement`
    strategies rank it in their top N (Req 12.2).
    """

    def __init__(
        self,
        min_agreement: int = _DEFAULT_CONSENSUS_MIN,
        top_n: int = 30,
    ) -> None:
        self._min_agreement = min_agreement
        self._top_n = top_n

    def rank(
        self,
        strategy_predictions: dict[str, pd.Series],
        weights: dict[str, float] | None = None,
    ) -> EnsembleResult:
        """Produce consensus-based stock ranking (Req 12.3).

        Args:
            strategy_predictions: Dict of strategy_name → Series(symbol → score).
            weights: Optional strategy weights (default: equal weight).

        Returns:
            EnsembleResult with ranked stocks meeting consensus threshold.
        """
        if not strategy_predictions:
            return EnsembleResult()

        n_strategies = len(strategy_predictions)
        if weights is None:
            weights = {name: 1.0 / n_strategies for name in strategy_predictions}

        # Collect top N from each strategy
        all_symbols: set[str] = set()
        strategy_tops: dict[str, set[str]] = {}

        for name, scores in strategy_predictions.items():
            top = scores.nlargest(self._top_n).index.tolist()
            strategy_tops[name] = set(top)
            all_symbols.update(top)

        # Score each symbol
        ranked: list[RankedStock] = []
        for symbol in all_symbols:
            agreement = sum(1 for tops in strategy_tops.values() if symbol in tops)

            # Only include if meets consensus threshold (Req 12.2)
            if agreement < self._min_agreement:
                continue

            # Weighted composite score
            strategy_scores = {}
            weighted_sum = 0.0
            weight_sum = 0.0
            for name, scores in strategy_predictions.items():
                if symbol in scores.index:
                    val = float(scores[symbol])
                    strategy_scores[name] = val
                    weighted_sum += val * weights.get(name, 1.0 / n_strategies)
                    weight_sum += weights.get(name, 1.0 / n_strategies)

            composite = weighted_sum / weight_sum if weight_sum > 0 else 0.0

            ranked.append(RankedStock(
                symbol=symbol,
                composite_score=composite,
                agreement_count=agreement,
                strategy_scores=strategy_scores,
            ))

        # Sort by composite score descending
        ranked.sort(key=lambda x: x.composite_score, reverse=True)

        return EnsembleResult(
            ranking=ranked,
            consensus_threshold=self._min_agreement,
            total_strategies=n_strategies,
        )


class AlphaDecayMonitor:
    """Monitors deployed signals for performance degradation (Req 12.4).

    Uses a 20-trading-day window for the Vietnamese market's faster
    alpha decay compared to developed markets (Req 12.5).
    """

    def __init__(self, window: int = _DECAY_WINDOW, ic_threshold: float = _DECAY_IC_THRESHOLD) -> None:
        self._window = window
        self._threshold = ic_threshold
        self._ic_history: dict[str, list[float]] = {}

    def update(self, signal_name: str, ic_value: float) -> DecayAlert | None:
        """Record a new IC observation and check for decay.

        Returns a DecayAlert if decay is detected, None otherwise.
        """
        if signal_name not in self._ic_history:
            self._ic_history[signal_name] = []

        history = self._ic_history[signal_name]
        history.append(ic_value)

        # Keep only recent window
        if len(history) > self._window * 3:
            self._ic_history[signal_name] = history[-self._window * 3:]
            history = self._ic_history[signal_name]

        if len(history) < self._window:
            return None

        recent = history[-self._window:]
        ic_current = np.mean(recent[-5:]) if len(recent) >= 5 else recent[-1]
        ic_peak = max(abs(x) for x in history)

        # Compute decay rate (slope over window)
        x = np.arange(len(recent))
        slope = float(np.polyfit(x, recent, 1)[0])

        # Decay detected if current IC below threshold and was previously strong
        if abs(ic_current) < self._threshold and ic_peak > self._threshold * 3:
            action = "replace" if slope < -0.002 else "rebalance"
            alert = DecayAlert(
                signal_name=signal_name,
                ic_current=float(ic_current),
                ic_peak=float(ic_peak),
                decay_rate=slope,
                action=action,
            )
            logger.warning(
                "Alpha decay detected: %s IC=%.4f (peak=%.4f, rate=%.4f) → %s",
                signal_name, ic_current, ic_peak, slope, action,
            )
            return alert

        return None

    def check_all(self) -> list[DecayAlert]:
        """Check all tracked signals for decay."""
        alerts = []
        for name, history in self._ic_history.items():
            if len(history) >= self._window:
                recent = history[-5:]
                ic_current = np.mean(recent)
                alert = self.update(name, ic_current)
                if alert:
                    alerts.append(alert)
        return alerts

    @property
    def tracked_signals(self) -> list[str]:
        return list(self._ic_history.keys())

    def get_ic_history(self, signal_name: str) -> list[float]:
        return self._ic_history.get(signal_name, [])
