"""Rolling accuracy computation per agent.

Tracks recommendation outcomes over 7d and 30d windows, computing
hit rates and directional accuracy for each agent.

Requirements: 41.2, 41.3, 41.5
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field
from typing import Any

logger = logging.getLogger(__name__)


@dataclass
class OutcomeRecord:
    """A single recommendation with its tracked outcome."""

    rec_id: str
    symbol: str
    agent_name: str
    action: str  # buy, sell, hold
    target_price: float
    confidence: int
    created_at: float  # Unix timestamp
    # Outcome fields (filled later)
    actual_return_1d: float | None = None
    actual_return_7d: float | None = None
    actual_return_14d: float | None = None
    actual_return_30d: float | None = None


@dataclass
class AgentAccuracy:
    """Accuracy metrics for a single agent."""

    agent_name: str
    accuracy_7d: float = 0.0   # % of correct directional calls over 7d
    accuracy_30d: float = 0.0  # % of correct directional calls over 30d
    total_recs_7d: int = 0
    total_recs_30d: int = 0
    avg_confidence_when_correct: float = 0.0
    avg_confidence_when_wrong: float = 0.0


class AccuracyTracker:
    """Computes rolling accuracy per agent from outcome records.

    A recommendation is "correct" if:
    - action=buy and actual_return > 0
    - action=sell and actual_return < 0
    - action=hold and abs(actual_return) < 3%
    """

    def __init__(self) -> None:
        self._records: list[OutcomeRecord] = []

    def add_record(self, record: OutcomeRecord) -> None:
        self._records.append(record)

    def load_records(self, records: list[OutcomeRecord]) -> None:
        self._records = records

    def compute_accuracy(self, agent_name: str, window_days: int = 30) -> AgentAccuracy:
        """Compute accuracy for a specific agent over a time window."""
        import time
        cutoff = time.time() - window_days * 86400

        relevant = [
            r for r in self._records
            if r.agent_name == agent_name and r.created_at >= cutoff
        ]

        if not relevant:
            return AgentAccuracy(agent_name=agent_name)

        # Use the appropriate return window
        correct = 0
        wrong = 0
        conf_correct: list[int] = []
        conf_wrong: list[int] = []

        for rec in relevant:
            ret = self._get_best_return(rec, window_days)
            if ret is None:
                continue

            is_correct = self._is_correct(rec.action, ret)
            if is_correct:
                correct += 1
                conf_correct.append(rec.confidence)
            else:
                wrong += 1
                conf_wrong.append(rec.confidence)

        total = correct + wrong
        accuracy = correct / total if total > 0 else 0.0

        return AgentAccuracy(
            agent_name=agent_name,
            accuracy_7d=accuracy if window_days <= 7 else 0.0,
            accuracy_30d=accuracy if window_days >= 30 else 0.0,
            total_recs_7d=total if window_days <= 7 else 0,
            total_recs_30d=total if window_days >= 30 else 0,
            avg_confidence_when_correct=sum(conf_correct) / len(conf_correct) if conf_correct else 0.0,
            avg_confidence_when_wrong=sum(conf_wrong) / len(conf_wrong) if conf_wrong else 0.0,
        )

    def compute_all_agents(self) -> dict[str, AgentAccuracy]:
        """Compute 30d accuracy for all agents."""
        agents = set(r.agent_name for r in self._records)
        return {name: self.compute_accuracy(name, 30) for name in agents}

    @staticmethod
    def _get_best_return(rec: OutcomeRecord, window: int) -> float | None:
        """Get the most relevant return for the given window."""
        if window <= 7 and rec.actual_return_7d is not None:
            return rec.actual_return_7d
        if window <= 14 and rec.actual_return_14d is not None:
            return rec.actual_return_14d
        if rec.actual_return_30d is not None:
            return rec.actual_return_30d
        if rec.actual_return_7d is not None:
            return rec.actual_return_7d
        return rec.actual_return_1d

    @staticmethod
    def _is_correct(action: str, actual_return: float) -> bool:
        if action == "buy":
            return actual_return > 0
        elif action == "sell":
            return actual_return < 0
        else:  # hold
            return abs(actual_return) < 0.03
