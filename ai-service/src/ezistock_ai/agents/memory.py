"""Agent Memory — bridges agents to persistent storage for long-term recall.

Connects the stateless agent pipeline to three backend storage systems:
1. Knowledge Base — market observations and pattern outcomes
2. Recommendation Tracker — full audit trail of every recommendation
3. Chat History — per-user conversation memory

This module handles read (before analysis) and write (after analysis) so
agents can learn from past recommendations and avoid repeating mistakes.

Requirements: 26.1, 26.2, 26.5, 32.1, 32.2, 41.4
"""

from __future__ import annotations

import logging
import time
from dataclasses import dataclass, field
from typing import Any

import httpx

from ezistock_ai.config import Config

logger = logging.getLogger(__name__)

_OUTCOME_INTERVALS = ("1d", "7d", "14d", "30d")


@dataclass
class PastRecommendation:
    """A previously stored recommendation for context injection."""

    symbol: str
    action: str
    target_price: float
    confidence_score: int
    created_at: str
    outcome: str = ""  # "correct", "incorrect", "pending"
    actual_return: float = 0.0


@dataclass
class PatternMatch:
    """A similar historical pattern from the Knowledge Base."""

    pattern_type: str
    symbol: str
    description: str
    confidence: float
    observed_at: str
    outcome: str
    accuracy: float = 0.0


@dataclass
class AgentMemoryContext:
    """Pre-fetched memory context injected into agents before each run."""

    past_recommendations: list[PastRecommendation] = field(default_factory=list)
    similar_patterns: list[PatternMatch] = field(default_factory=list)
    agent_accuracy: dict[str, float] = field(default_factory=dict)  # agent_name → 30d accuracy
    accuracy_summary: str = ""  # Human-readable summary for prompt injection

    def to_prompt_context(self) -> str:
        """Format memory context as a string for injection into agent prompts."""
        parts: list[str] = []

        if self.past_recommendations:
            recs = self.past_recommendations[:5]
            rec_lines = []
            for r in recs:
                outcome_str = f" → {r.outcome}" if r.outcome else " (pending)"
                rec_lines.append(
                    f"  - {r.created_at}: {r.action.upper()} {r.symbol} "
                    f"@ {r.target_price:,.0f} (conf={r.confidence_score}){outcome_str}"
                )
            parts.append("Recent recommendations:\n" + "\n".join(rec_lines))

        if self.similar_patterns:
            pat_lines = []
            for p in self.similar_patterns[:3]:
                pat_lines.append(
                    f"  - {p.pattern_type} on {p.symbol} ({p.observed_at}): "
                    f"{p.description} → {p.outcome} (accuracy={p.accuracy:.0%})"
                )
            parts.append("Similar historical patterns:\n" + "\n".join(pat_lines))

        if self.agent_accuracy:
            acc_lines = [f"  - {name}: {acc:.0%}" for name, acc in self.agent_accuracy.items()]
            parts.append("Agent 30-day accuracy:\n" + "\n".join(acc_lines))

        if self.accuracy_summary:
            parts.append(f"Accuracy note: {self.accuracy_summary}")

        return "\n\n".join(parts) if parts else "No historical data available."


class AgentMemory:
    """Reads from and writes to the Go backend's persistent stores.

    All calls go through the Go backend REST API, which owns the PostgreSQL
    database. The Python AI service never touches the DB directly.
    """

    def __init__(self, config: Config | None = None) -> None:
        self._config = config or Config()
        self._backend_url = self._config.go_backend_url
        self._timeout = 5.0  # Fast timeout — memory is best-effort

    # ------------------------------------------------------------------
    # READ: Pre-fetch context before agent run
    # ------------------------------------------------------------------

    async def load_context(
        self,
        symbol: str,
        user_id: str = "",
        pattern_type: str = "",
    ) -> AgentMemoryContext:
        """Load all relevant memory context for a symbol before analysis.

        This is called by the orchestrator before running agents, and the
        result is injected into AgentContext.accuracy_context.
        """
        ctx = AgentMemoryContext()

        # Fetch in parallel (best-effort, failures are non-fatal)
        import asyncio
        results = await asyncio.gather(
            self._fetch_past_recommendations(symbol),
            self._fetch_similar_patterns(symbol, pattern_type),
            self._fetch_agent_accuracy(),
            return_exceptions=True,
        )

        if isinstance(results[0], list):
            ctx.past_recommendations = results[0]
        else:
            logger.debug("Failed to fetch past recommendations: %s", results[0])

        if isinstance(results[1], list):
            ctx.similar_patterns = results[1]
        else:
            logger.debug("Failed to fetch similar patterns: %s", results[1])

        if isinstance(results[2], dict):
            ctx.agent_accuracy = results[2]
        else:
            logger.debug("Failed to fetch agent accuracy: %s", results[2])

        # Build human-readable summary
        ctx.accuracy_summary = self._build_accuracy_summary(ctx)

        return ctx

    async def _fetch_past_recommendations(self, symbol: str) -> list[PastRecommendation]:
        """Fetch recent recommendations for a symbol from the tracker."""
        url = f"{self._backend_url}/api/ranking/recommendations?symbol={symbol}&limit=10"
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            resp = await client.get(url)
            if resp.status_code != 200:
                return []
            data = resp.json()
            return [
                PastRecommendation(
                    symbol=item.get("symbol", symbol),
                    action=item.get("action", ""),
                    target_price=item.get("target_price", 0),
                    confidence_score=item.get("confidence_score", 0),
                    created_at=item.get("created_at", ""),
                    outcome=item.get("outcome", ""),
                    actual_return=item.get("actual_return", 0),
                )
                for item in data.get("recommendations", [])
            ]

    async def _fetch_similar_patterns(self, symbol: str, pattern_type: str) -> list[PatternMatch]:
        """Query Knowledge Base for similar historical patterns."""
        url = f"{self._backend_url}/api/knowledge/patterns?symbol={symbol}&min_confidence=0.5"
        if pattern_type:
            url += f"&pattern_type={pattern_type}"
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            resp = await client.get(url)
            if resp.status_code != 200:
                return []
            data = resp.json()
            return [
                PatternMatch(
                    pattern_type=item.get("pattern_type", ""),
                    symbol=item.get("symbol", symbol),
                    description=item.get("description", ""),
                    confidence=item.get("confidence", 0),
                    observed_at=item.get("observed_at", ""),
                    outcome=item.get("outcome", ""),
                    accuracy=item.get("accuracy", 0),
                )
                for item in data.get("patterns", [])
            ]

    async def _fetch_agent_accuracy(self) -> dict[str, float]:
        """Fetch per-agent accuracy from the Feedback Loop."""
        url = f"{self._backend_url}/api/feedback/accuracy"
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            resp = await client.get(url)
            if resp.status_code != 200:
                return {}
            return resp.json().get("accuracy", {})

    # ------------------------------------------------------------------
    # WRITE: Persist results after agent run
    # ------------------------------------------------------------------

    async def store_recommendation(
        self,
        symbol: str,
        action: str,
        target_price: float,
        confidence_score: int,
        risk_level: str,
        reasoning: str,
        agent_outputs: dict[str, Any] | None = None,
    ) -> bool:
        """Persist a recommendation to the tracker for outcome tracking."""
        url = f"{self._backend_url}/api/ranking/recommendations"
        payload = {
            "symbol": symbol,
            "action": action,
            "target_price": target_price,
            "confidence_score": confidence_score,
            "risk_level": risk_level,
            "reasoning": reasoning,
            "agent_outputs": agent_outputs or {},
            "created_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        }
        try:
            async with httpx.AsyncClient(timeout=self._timeout) as client:
                resp = await client.post(url, json=payload)
                if resp.status_code in (200, 201):
                    logger.info("Stored recommendation for %s: %s", symbol, action)
                    return True
                logger.warning("Failed to store recommendation: %d", resp.status_code)
                return False
        except Exception as exc:
            logger.warning("Failed to store recommendation for %s: %s", symbol, exc)
            return False

    async def store_observation(
        self,
        symbol: str,
        pattern_type: str,
        description: str,
        confidence: float,
    ) -> bool:
        """Record a market observation in the Knowledge Base."""
        url = f"{self._backend_url}/api/knowledge/observations"
        payload = {
            "symbol": symbol,
            "pattern_type": pattern_type,
            "description": description,
            "confidence": confidence,
            "observed_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        }
        try:
            async with httpx.AsyncClient(timeout=self._timeout) as client:
                resp = await client.post(url, json=payload)
                return resp.status_code in (200, 201)
        except Exception as exc:
            logger.debug("Failed to store observation: %s", exc)
            return False

    # ------------------------------------------------------------------
    # Chat history
    # ------------------------------------------------------------------

    async def load_chat_history(self, user_id: str, limit: int = 20) -> list[dict[str, str]]:
        """Load recent chat messages for a user."""
        url = f"{self._backend_url}/api/chat/history?user_id={user_id}&limit={limit}"
        try:
            async with httpx.AsyncClient(timeout=self._timeout) as client:
                resp = await client.get(url)
                if resp.status_code != 200:
                    return []
                return resp.json().get("messages", [])
        except Exception:
            return []

    async def store_chat_message(
        self,
        user_id: str,
        role: str,
        content: str,
    ) -> bool:
        """Persist a chat message."""
        url = f"{self._backend_url}/api/chat/history"
        try:
            async with httpx.AsyncClient(timeout=self._timeout) as client:
                resp = await client.post(url, json={
                    "user_id": user_id,
                    "role": role,
                    "content": content,
                    "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                })
                return resp.status_code in (200, 201)
        except Exception:
            return False

    # ------------------------------------------------------------------
    # Helpers
    # ------------------------------------------------------------------

    @staticmethod
    def _build_accuracy_summary(ctx: AgentMemoryContext) -> str:
        """Build a one-liner about historical accuracy for prompt injection."""
        if not ctx.past_recommendations:
            return ""

        correct = sum(1 for r in ctx.past_recommendations if r.outcome == "correct")
        total_with_outcome = sum(1 for r in ctx.past_recommendations if r.outcome in ("correct", "incorrect"))

        if total_with_outcome == 0:
            return f"Last {len(ctx.past_recommendations)} recommendations pending outcome tracking."

        accuracy = correct / total_with_outcome
        return (
            f"Historical accuracy for this symbol: {accuracy:.0%} "
            f"({correct}/{total_with_outcome} correct). "
            f"Calibrate confidence accordingly."
        )
