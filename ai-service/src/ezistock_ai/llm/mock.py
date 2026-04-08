"""Mock LLM for deterministic E2E and integration tests.

When EZISTOCK_TEST_MODE=true, the LLMRouter returns MockChatModel instances
that produce canned responses without calling any external LLM provider.

Requirements: 46.3
"""

from __future__ import annotations

import json
from typing import Any, Iterator, List, Optional

from langchain_core.callbacks import CallbackManagerForLLMRun
from langchain_core.language_models import BaseChatModel
from langchain_core.messages import AIMessage, BaseMessage
from langchain_core.outputs import ChatGeneration, ChatResult


class MockChatModel(BaseChatModel):
    """A fake LLM that returns deterministic responses for testing.

    Responses are selected based on keywords in the input message.
    """

    model_name: str = "mock-model"

    @property
    def _llm_type(self) -> str:
        return "mock"

    def _generate(
        self,
        messages: List[BaseMessage],
        stop: Optional[List[str]] = None,
        run_manager: Optional[CallbackManagerForLLMRun] = None,
        **kwargs: Any,
    ) -> ChatResult:
        last_msg = messages[-1].content if messages else ""
        content = _select_response(str(last_msg))
        return ChatResult(generations=[ChatGeneration(message=AIMessage(content=content))])


# ---------------------------------------------------------------------------
# Canned responses keyed by topic detection
# ---------------------------------------------------------------------------

_TECHNICAL_RESPONSE = json.dumps({
    "composite_signal": "bullish",
    "indicators": {"rsi": 55.2, "macd": 0.5, "adx": 28.1},
    "support_levels": [95000, 92000],
    "resistance_levels": [102000, 105000],
    "patterns": ["hammer"],
    "smart_money_flow": "moderate_inflow",
})

_NEWS_RESPONSE = json.dumps({
    "sentiment": "positive",
    "confidence": 0.72,
    "catalysts": ["Strong Q4 earnings", "New product launch"],
    "risk_factors": ["Rising input costs"],
    "articles": [
        {"title": "FPT reports record revenue", "url": "https://example.com/1", "source": "CafeF"},
    ],
})

_RECOMMENDATION_RESPONSE = json.dumps({
    "action": "buy",
    "target_price": 105000,
    "upside_percent": 10.5,
    "confidence_score": 72,
    "risk_level": "medium",
    "reasoning": "Strong technical setup with bullish momentum and positive news catalysts.",
})

_STRATEGY_RESPONSE = json.dumps({
    "signal_direction": "long",
    "entry_price": 95000,
    "stop_loss": 91000,
    "take_profit": 105000,
    "risk_reward_ratio": 2.5,
    "position_size_percent": 5.0,
    "reasoning": "ATR-based entry with favorable risk/reward.",
})

_DEFAULT_RESPONSE = "This is a mock AI response for testing purposes."


def _select_response(text: str) -> str:
    """Select a canned response based on keywords in the input."""
    lower = text.lower()
    if any(kw in lower for kw in ("technical", "indicator", "rsi", "macd")):
        return _TECHNICAL_RESPONSE
    if any(kw in lower for kw in ("news", "sentiment", "catalyst")):
        return _NEWS_RESPONSE
    if any(kw in lower for kw in ("recommend", "advisor", "investment")):
        return _RECOMMENDATION_RESPONSE
    if any(kw in lower for kw in ("strategy", "entry", "stop_loss", "trade")):
        return _STRATEGY_RESPONSE
    return _DEFAULT_RESPONSE
