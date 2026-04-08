"""LLM cost tracking — per-user/agent/model token usage with daily budget enforcement.

Requirements: 48.1, 48.2, 48.4, 48.6
"""

from __future__ import annotations

import logging
import time
from dataclasses import dataclass, field
from typing import Optional

from ezistock_ai.config import Config

logger = logging.getLogger(__name__)


@dataclass
class UsageRecord:
    """A single LLM call usage record."""

    user_id: str
    agent_name: str
    model: str
    input_tokens: int
    output_tokens: int
    cost_usd: float
    timestamp: float  # Unix epoch


@dataclass
class UserBudget:
    """Tracks a user's daily token consumption."""

    user_id: str
    tokens_used: int = 0
    cost_usd: float = 0.0
    day_key: str = ""  # YYYY-MM-DD, reset when day changes


class BudgetExceededError(Exception):
    """Raised when a user exceeds their daily token budget."""

    def __init__(self, user_id: str, used: int, limit: int) -> None:
        self.user_id = user_id
        self.used = used
        self.limit = limit
        super().__init__(f"User {user_id} exceeded daily budget: {used}/{limit} tokens")


@dataclass
class CostSummary:
    """Aggregated cost summary for reporting."""

    total_tokens: int = 0
    total_cost_usd: float = 0.0
    by_agent: dict[str, float] = field(default_factory=dict)
    by_model: dict[str, float] = field(default_factory=dict)
    by_user: dict[str, float] = field(default_factory=dict)


class CostTracker:
    """Tracks token usage and enforces daily budgets.

    In-memory implementation suitable for single-process deployment.
    For multi-process, swap the backing store to Redis or PostgreSQL.
    """

    # Default budget tiers (tokens/day)
    FREE_TIER_LIMIT = 100_000
    PREMIUM_TIER_LIMIT = 500_000

    def __init__(self, config: Config) -> None:
        self._config = config
        self._daily_budget_usd = config.llm_daily_budget_usd
        self._records: list[UsageRecord] = []
        self._user_budgets: dict[str, UserBudget] = {}

    def record_usage(
        self,
        user_id: str,
        agent_name: str,
        model: str,
        input_tokens: int,
        output_tokens: int,
        cost_usd: float,
    ) -> UsageRecord:
        """Record a completed LLM call's token usage."""
        record = UsageRecord(
            user_id=user_id,
            agent_name=agent_name,
            model=model,
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            cost_usd=cost_usd,
            timestamp=time.time(),
        )
        self._records.append(record)

        # Update user budget tracking
        budget = self._get_or_create_budget(user_id)
        budget.tokens_used += input_tokens + output_tokens
        budget.cost_usd += cost_usd

        logger.debug(
            "LLM usage: user=%s agent=%s model=%s tokens=%d+%d cost=$%.6f",
            user_id, agent_name, model, input_tokens, output_tokens, cost_usd,
        )
        return record

    def check_budget(self, user_id: str, token_limit: Optional[int] = None) -> None:
        """Check if user is within daily budget. Raises BudgetExceededError if not.

        Args:
            user_id: The user to check.
            token_limit: Override token limit. Defaults to FREE_TIER_LIMIT.
        """
        limit = token_limit or self.FREE_TIER_LIMIT
        budget = self._get_or_create_budget(user_id)
        if budget.tokens_used >= limit:
            raise BudgetExceededError(user_id, budget.tokens_used, limit)

    def is_near_budget(self, user_id: str, threshold: float = 0.8, token_limit: Optional[int] = None) -> bool:
        """Return True if user has consumed >= threshold of their daily budget."""
        limit = token_limit or self.FREE_TIER_LIMIT
        budget = self._get_or_create_budget(user_id)
        return budget.tokens_used >= int(limit * threshold)

    def get_user_usage(self, user_id: str) -> tuple[int, float]:
        """Return (tokens_used_today, cost_usd_today) for a user."""
        budget = self._get_or_create_budget(user_id)
        return budget.tokens_used, budget.cost_usd

    def get_daily_summary(self) -> CostSummary:
        """Aggregate today's usage across all users/agents/models."""
        today = _today_key()
        summary = CostSummary()
        for rec in self._records:
            if _day_key(rec.timestamp) != today:
                continue
            total = rec.input_tokens + rec.output_tokens
            summary.total_tokens += total
            summary.total_cost_usd += rec.cost_usd
            summary.by_agent[rec.agent_name] = summary.by_agent.get(rec.agent_name, 0.0) + rec.cost_usd
            summary.by_model[rec.model] = summary.by_model.get(rec.model, 0.0) + rec.cost_usd
            summary.by_user[rec.user_id] = summary.by_user.get(rec.user_id, 0.0) + rec.cost_usd
        return summary

    def estimate_cost(self, model: str, input_tokens: int, output_tokens: int) -> float:
        """Estimate cost in USD for a planned LLM call."""
        from ezistock_ai.llm.router import MODEL_COSTS

        input_rate, output_rate = MODEL_COSTS.get(model, (0.01, 0.03))
        return (input_tokens / 1000) * input_rate + (output_tokens / 1000) * output_rate

    # ------------------------------------------------------------------
    # Private helpers
    # ------------------------------------------------------------------

    def _get_or_create_budget(self, user_id: str) -> UserBudget:
        today = _today_key()
        budget = self._user_budgets.get(user_id)
        if budget is None or budget.day_key != today:
            budget = UserBudget(user_id=user_id, day_key=today)
            self._user_budgets[user_id] = budget
        return budget


def _today_key() -> str:
    return time.strftime("%Y-%m-%d", time.gmtime())


def _day_key(ts: float) -> str:
    return time.strftime("%Y-%m-%d", time.gmtime(ts))
