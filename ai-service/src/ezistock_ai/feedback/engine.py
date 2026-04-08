"""Feedback Loop Engine — orchestrates recommendation → outcome → improvement.

Closes the loop between what agents recommend and what actually happens.
Computes per-agent accuracy, detects biases, adjusts agent weights,
and injects historical performance context into agent prompts.

Requirements: 41.1, 41.2, 41.3, 41.4, 41.5, 41.6, 41.7, 41.9, 41.10
"""

from __future__ import annotations

import logging
import time
from dataclasses import dataclass, field

import httpx

from ezistock_ai.config import Config
from ezistock_ai.feedback.accuracy import AccuracyTracker, AgentAccuracy, OutcomeRecord
from ezistock_ai.feedback.bias import BiasDetector, BiasReport

logger = logging.getLogger(__name__)

# Weight reduction thresholds (Req 41.9)
_WEIGHT_REDUCE_THRESHOLD = 0.40  # Reduce weight if 30d accuracy < 40%
_WEIGHT_RESTORE_THRESHOLD = 0.50  # Restore weight if accuracy > 50%
_REDUCED_WEIGHT = 0.5  # Multiplier when agent is in reduced-weight mode
_NORMAL_WEIGHT = 1.0


@dataclass
class AgentWeight:
    """Dynamic weight for an agent based on accuracy."""

    agent_name: str
    weight: float = _NORMAL_WEIGHT
    reduced: bool = False
    accuracy_30d: float = 0.0
    last_updated: float = 0.0


@dataclass
class FeedbackContext:
    """Full feedback context for injection into the agent pipeline."""

    agent_accuracies: dict[str, AgentAccuracy] = field(default_factory=dict)
    agent_biases: dict[str, BiasReport] = field(default_factory=dict)
    agent_weights: dict[str, AgentWeight] = field(default_factory=dict)
    accuracy_prompt: str = ""  # Combined prompt injection string
    bias_prompt: str = ""      # Combined bias correction string

    def to_prompt_context(self) -> str:
        """Format for injection into agent prompts."""
        parts = []
        if self.accuracy_prompt:
            parts.append(self.accuracy_prompt)
        if self.bias_prompt:
            parts.append(self.bias_prompt)
        return "\n".join(parts) if parts else ""


class FeedbackEngine:
    """Orchestrates the full feedback loop.

    Flow:
    1. Fetch outcome data from Go backend (recommendation tracker)
    2. Compute per-agent accuracy (7d, 30d rolling windows)
    3. Detect systematic biases
    4. Adjust agent weights
    5. Generate context injection strings for agent prompts
    6. Feed outcomes back to Alpha Mining Engine for model retraining
    """

    def __init__(self, config: Config | None = None) -> None:
        self._config = config or Config()
        self._backend_url = self._config.go_backend_url
        self._accuracy = AccuracyTracker()
        self._bias = BiasDetector()
        self._weights: dict[str, AgentWeight] = {}
        self._agent_names = [
            "technical_analyst", "news_analyst",
            "investment_advisor", "strategy_builder",
        ]

    async def run(self) -> FeedbackContext:
        """Execute the full feedback loop. Called after outcome tracking updates.

        Returns FeedbackContext that should be injected into the next agent run.
        """
        # Step 1: Fetch outcome records from Go backend
        records = await self._fetch_outcomes()
        if not records:
            logger.info("No outcome records available for feedback loop")
            return FeedbackContext()

        # Step 2: Compute accuracy
        self._accuracy.load_records(records)
        accuracies = self._accuracy.compute_all_agents()

        # Step 3: Detect biases
        self._bias.load_records(records)
        biases = {}
        for name in self._agent_names:
            biases[name] = self._bias.detect(name)

        # Step 4: Adjust weights (Req 41.9)
        self._update_weights(accuracies)

        # Step 5: Generate prompt context
        accuracy_prompt = self._build_accuracy_prompt(accuracies)
        bias_prompt = self._build_bias_prompt(biases)

        ctx = FeedbackContext(
            agent_accuracies=accuracies,
            agent_biases=biases,
            agent_weights=dict(self._weights),
            accuracy_prompt=accuracy_prompt,
            bias_prompt=bias_prompt,
        )

        logger.info(
            "Feedback loop complete: %d records, agents=%s",
            len(records),
            {n: f"{a.accuracy_30d:.0%}" for n, a in accuracies.items()},
        )

        return ctx

    def get_agent_weight(self, agent_name: str) -> float:
        """Get current weight for an agent (used by orchestrator)."""
        w = self._weights.get(agent_name)
        return w.weight if w else _NORMAL_WEIGHT

    def get_weights(self) -> dict[str, float]:
        """Get all agent weights."""
        return {name: self.get_agent_weight(name) for name in self._agent_names}

    # ------------------------------------------------------------------
    # Data fetching
    # ------------------------------------------------------------------

    async def _fetch_outcomes(self) -> list[OutcomeRecord]:
        """Fetch recommendation outcomes from Go backend."""
        url = f"{self._backend_url}/api/feedback/outcomes?days=60"
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                resp = await client.get(url)
                if resp.status_code != 200:
                    return []
                data = resp.json()
                return [
                    OutcomeRecord(
                        rec_id=item.get("id", ""),
                        symbol=item.get("symbol", ""),
                        agent_name=item.get("agent_name", ""),
                        action=item.get("action", ""),
                        target_price=item.get("target_price", 0),
                        confidence=item.get("confidence_score", 0),
                        created_at=item.get("created_at_unix", 0),
                        actual_return_1d=item.get("return_1d"),
                        actual_return_7d=item.get("return_7d"),
                        actual_return_14d=item.get("return_14d"),
                        actual_return_30d=item.get("return_30d"),
                    )
                    for item in data.get("outcomes", [])
                ]
        except Exception as exc:
            logger.warning("Failed to fetch outcomes: %s", exc)
            return []

    # ------------------------------------------------------------------
    # Weight management (Req 41.9)
    # ------------------------------------------------------------------

    def _update_weights(self, accuracies: dict[str, AgentAccuracy]) -> None:
        """Adjust agent weights based on 30d accuracy."""
        for name in self._agent_names:
            acc = accuracies.get(name)
            if not acc or acc.total_recs_30d < 5:
                continue

            w = self._weights.get(name, AgentWeight(agent_name=name))

            if acc.accuracy_30d < _WEIGHT_REDUCE_THRESHOLD and not w.reduced:
                w.weight = _REDUCED_WEIGHT
                w.reduced = True
                logger.warning(
                    "Reducing weight for %s: accuracy=%.0f%% < %.0f%%",
                    name, acc.accuracy_30d * 100, _WEIGHT_REDUCE_THRESHOLD * 100,
                )
            elif acc.accuracy_30d >= _WEIGHT_RESTORE_THRESHOLD and w.reduced:
                w.weight = _NORMAL_WEIGHT
                w.reduced = False
                logger.info(
                    "Restoring weight for %s: accuracy=%.0f%% >= %.0f%%",
                    name, acc.accuracy_30d * 100, _WEIGHT_RESTORE_THRESHOLD * 100,
                )

            w.accuracy_30d = acc.accuracy_30d
            w.last_updated = time.time()
            self._weights[name] = w

    # ------------------------------------------------------------------
    # Prompt generation
    # ------------------------------------------------------------------

    @staticmethod
    def _build_accuracy_prompt(accuracies: dict[str, AgentAccuracy]) -> str:
        lines = ["Historical agent accuracy (30-day rolling):"]
        for name, acc in accuracies.items():
            if acc.total_recs_30d > 0:
                lines.append(
                    f"  {name}: {acc.accuracy_30d:.0%} ({acc.total_recs_30d} recs, "
                    f"avg conf correct={acc.avg_confidence_when_correct:.0f}, "
                    f"wrong={acc.avg_confidence_when_wrong:.0f})"
                )
        return "\n".join(lines) if len(lines) > 1 else ""

    @staticmethod
    def _build_bias_prompt(biases: dict[str, BiasReport]) -> str:
        corrections = [b.correction_prompt for b in biases.values() if b.correction_prompt]
        return " ".join(corrections)
