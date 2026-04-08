"""Systematic bias detection and correction.

Detects patterns like sector overestimation, timing errors, and
overconfidence. Generates correction factors that get injected
into agent prompts.

Requirements: 41.4, 41.6, 41.7
"""

from __future__ import annotations

import logging
from collections import defaultdict
from dataclasses import dataclass, field

from ezistock_ai.feedback.accuracy import OutcomeRecord

logger = logging.getLogger(__name__)


@dataclass
class BiasReport:
    """Detected biases for an agent."""

    agent_name: str
    sector_biases: dict[str, float] = field(default_factory=dict)  # sector → avg error
    timing_bias: float = 0.0  # Positive = too early, negative = too late
    confidence_bias: float = 0.0  # Positive = overconfident, negative = underconfident
    direction_bias: float = 0.0  # Positive = bullish bias, negative = bearish bias
    correction_prompt: str = ""  # Injected into agent prompts


class BiasDetector:
    """Detects systematic biases from outcome records."""

    def __init__(self) -> None:
        self._records: list[OutcomeRecord] = []

    def load_records(self, records: list[OutcomeRecord]) -> None:
        self._records = records

    def detect(self, agent_name: str) -> BiasReport:
        """Analyze an agent's track record for systematic biases."""
        relevant = [r for r in self._records if r.agent_name == agent_name]
        if len(relevant) < 10:
            return BiasReport(agent_name=agent_name)

        report = BiasReport(agent_name=agent_name)

        # Sector bias: does the agent consistently overestimate certain sectors?
        report.sector_biases = self._detect_sector_bias(relevant)

        # Direction bias: does the agent lean bullish or bearish?
        report.direction_bias = self._detect_direction_bias(relevant)

        # Confidence bias: is the agent overconfident or underconfident?
        report.confidence_bias = self._detect_confidence_bias(relevant)

        # Generate correction prompt
        report.correction_prompt = self._generate_correction_prompt(report)

        return report

    def _detect_sector_bias(self, records: list[OutcomeRecord]) -> dict[str, float]:
        """Group by sector (from symbol prefix) and compute avg prediction error."""
        # Simple heuristic: group by first letter of symbol as proxy for exchange
        # In production, this would use actual ICB sector mapping
        sector_errors: dict[str, list[float]] = defaultdict(list)

        for rec in records:
            ret = rec.actual_return_7d or rec.actual_return_1d
            if ret is None:
                continue

            # Compute prediction error: expected direction vs actual
            expected = 1.0 if rec.action == "buy" else (-1.0 if rec.action == "sell" else 0.0)
            actual_dir = 1.0 if ret > 0 else (-1.0 if ret < 0 else 0.0)
            error = expected - actual_dir

            # Use symbol as sector proxy (will be replaced with real sector mapping)
            sector = rec.symbol[:2] if len(rec.symbol) >= 2 else "XX"
            sector_errors[sector].append(error)

        return {
            sector: sum(errors) / len(errors)
            for sector, errors in sector_errors.items()
            if len(errors) >= 3 and abs(sum(errors) / len(errors)) > 0.3
        }

    def _detect_direction_bias(self, records: list[OutcomeRecord]) -> float:
        """Check if agent leans bullish or bearish overall."""
        buy_count = sum(1 for r in records if r.action == "buy")
        sell_count = sum(1 for r in records if r.action == "sell")
        total = buy_count + sell_count
        if total == 0:
            return 0.0
        return (buy_count - sell_count) / total  # +1 = all buy, -1 = all sell

    def _detect_confidence_bias(self, records: list[OutcomeRecord]) -> float:
        """Check if high-confidence calls are actually more accurate."""
        high_conf = [r for r in records if r.confidence >= 70]
        low_conf = [r for r in records if r.confidence < 50]

        high_correct = self._hit_rate(high_conf)
        low_correct = self._hit_rate(low_conf)

        # If high-confidence calls aren't more accurate, agent is overconfident
        if high_correct is not None and low_correct is not None:
            return high_correct - low_correct  # Negative = overconfident
        return 0.0

    @staticmethod
    def _hit_rate(records: list[OutcomeRecord]) -> float | None:
        if not records:
            return None
        correct = 0
        total = 0
        for r in records:
            ret = r.actual_return_7d or r.actual_return_1d
            if ret is None:
                continue
            total += 1
            if r.action == "buy" and ret > 0:
                correct += 1
            elif r.action == "sell" and ret < 0:
                correct += 1
        return correct / total if total > 0 else None

    @staticmethod
    def _generate_correction_prompt(report: BiasReport) -> str:
        """Generate a correction string to inject into agent prompts."""
        corrections: list[str] = []

        if report.direction_bias > 0.3:
            corrections.append(
                f"BIAS WARNING: You have a bullish bias ({report.direction_bias:.0%} net buy). "
                "Be more critical of buy signals and give extra weight to bearish indicators."
            )
        elif report.direction_bias < -0.3:
            corrections.append(
                f"BIAS WARNING: You have a bearish bias ({report.direction_bias:.0%} net sell). "
                "Be more open to bullish setups and recovery patterns."
            )

        if report.confidence_bias < -0.15:
            corrections.append(
                "CALIBRATION: Your high-confidence calls aren't more accurate than low-confidence ones. "
                "Lower your confidence scores by 10-15 points."
            )

        for sector, error in list(report.sector_biases.items())[:3]:
            if error > 0.3:
                corrections.append(f"SECTOR BIAS: You overestimate {sector} stocks. Apply extra skepticism.")
            elif error < -0.3:
                corrections.append(f"SECTOR BIAS: You underestimate {sector} stocks. Consider more upside.")

        return " ".join(corrections)
