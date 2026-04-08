"""LLM model routing — selects the cheapest adequate model per task type.

Lightweight models (e.g. GPT-4o-mini, Claude Haiku) for data extraction/summarization.
Capable models (e.g. GPT-4o, Claude Sonnet) for analysis/recommendations.
Conversational models for user-facing chat responses.

Requirements: 4.3, 4.7, 48.3
"""

from __future__ import annotations

import logging
from enum import Enum
from typing import Any

from langchain_core.language_models import BaseChatModel

from ezistock_ai.config import Config

logger = logging.getLogger(__name__)


class TaskType(str, Enum):
    """Classification of LLM tasks for model routing."""

    EXTRACTION = "extraction"  # Data extraction, classification, summarization
    ANALYSIS = "analysis"  # Deep analysis, synthesis, recommendations
    CONVERSATION = "conversation"  # User-facing chat responses


# Approximate cost per 1K tokens (input, output) in USD — used for estimation.
MODEL_COSTS: dict[str, tuple[float, float]] = {
    # OpenAI
    "gpt-4o-mini": (0.00015, 0.0006),
    "gpt-4o": (0.0025, 0.01),
    "gpt-4-turbo": (0.01, 0.03),
    # Anthropic
    "claude-3-haiku-20240307": (0.00025, 0.00125),
    "claude-3-5-haiku-20241022": (0.0008, 0.004),
    "claude-3-5-sonnet-20241022": (0.003, 0.015),
    "claude-3-opus-20240229": (0.015, 0.075),
    # Google
    "gemini-1.5-flash": (0.000075, 0.0003),
    "gemini-1.5-pro": (0.00125, 0.005),
}


class LLMRouter:
    """Routes LLM requests to the appropriate model based on task type.

    Uses config-driven model names so operators can swap models without code changes.
    """

    def __init__(self, config: Config) -> None:
        self._config = config
        self._model_map: dict[TaskType, str] = {
            TaskType.EXTRACTION: config.llm_lightweight_model,
            TaskType.ANALYSIS: config.llm_capable_model,
            TaskType.CONVERSATION: config.llm_conversational_model,
        }
        self._instances: dict[str, BaseChatModel] = {}

    def get_model_name(self, task_type: TaskType) -> str:
        """Return the configured model name for a task type."""
        return self._model_map[task_type]

    def get_model(self, task_type: TaskType) -> BaseChatModel:
        """Return a LangChain chat model instance for the given task type.

        Instances are lazily created and cached for reuse.
        """
        model_name = self._model_map[task_type]
        if model_name not in self._instances:
            self._instances[model_name] = self._create_model(model_name)
            logger.info("Created LLM instance: %s for task_type=%s", model_name, task_type.value)
        return self._instances[model_name]

    def get_cost_per_1k(self, model_name: str) -> tuple[float, float]:
        """Return (input_cost, output_cost) per 1K tokens for a model.

        Falls back to a conservative default if the model is unknown.
        """
        return MODEL_COSTS.get(model_name, (0.01, 0.03))

    # ------------------------------------------------------------------
    # Private helpers
    # ------------------------------------------------------------------

    def _create_model(self, model_name: str) -> BaseChatModel:
        """Instantiate the correct LangChain provider based on model name.

        In test mode (config.test_mode=True), returns a MockChatModel
        that produces deterministic responses without external API calls.
        Requirements: 46.3
        """
        if self._config.test_mode:
            from ezistock_ai.llm.mock import MockChatModel

            logger.info("TEST MODE: using MockChatModel for %s", model_name)
            return MockChatModel(model_name=f"mock-{model_name}")

        kwargs: dict[str, Any] = {"model": model_name, "temperature": 0.3}

        if model_name.startswith("gpt-") or model_name.startswith("o1") or model_name.startswith("o3"):
            return self._create_openai(model_name, kwargs)
        if model_name.startswith("claude-"):
            return self._create_anthropic(model_name, kwargs)
        if model_name.startswith("gemini-"):
            return self._create_google(model_name, kwargs)

        # Default to OpenAI-compatible
        logger.warning("Unknown model prefix for '%s', defaulting to OpenAI provider", model_name)
        return self._create_openai(model_name, kwargs)

    def _create_openai(self, model_name: str, kwargs: dict[str, Any]) -> BaseChatModel:
        from langchain_openai import ChatOpenAI

        return ChatOpenAI(api_key=self._config.openai_api_key, **kwargs)  # type: ignore[arg-type]

    def _create_anthropic(self, model_name: str, kwargs: dict[str, Any]) -> BaseChatModel:
        from langchain_anthropic import ChatAnthropic

        return ChatAnthropic(api_key=self._config.anthropic_api_key, **kwargs)  # type: ignore[arg-type]

    def _create_google(self, model_name: str, kwargs: dict[str, Any]) -> BaseChatModel:
        from langchain_google_genai import ChatGoogleGenerativeAI

        return ChatGoogleGenerativeAI(google_api_key=self._config.google_api_key, **kwargs)  # type: ignore[arg-type]
