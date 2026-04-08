"""BaseAgent — shared interface for all multi-agent system agents.

Provides consistent timeout enforcement (30s per Req 4.6), structured logging,
error handling, citation collection, and LLM cache integration.

Requirements: 4.1, 4.6, 4.8, 48.4
"""

from __future__ import annotations

import asyncio
import logging
import time
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Any, Generic, TypeVar

from ezistock_ai.generated.proto.agent_pb2 import Citation, MarketData

logger = logging.getLogger(__name__)

T = TypeVar("T")  # Agent output type

DEFAULT_TIMEOUT_SECONDS = 30.0


@dataclass
class AgentContext:
    """Shared context passed to every agent invocation."""

    symbol: str
    market_data: MarketData
    user_id: str = ""
    # Optional enrichment data (populated by context_enricher before run)
    historical_accuracy: float | None = None
    accuracy_context: str = ""  # Injected into prompts for Feedback Loop (Req 32)
    extra: dict[str, Any] = field(default_factory=dict)


@dataclass
class AgentResult(Generic[T]):
    """Wrapper around an agent's output with metadata."""

    agent_name: str
    output: T | None = None
    citations: list[Citation] = field(default_factory=list)
    elapsed_seconds: float = 0.0
    success: bool = True
    error: str | None = None
    # Degradation metadata (item 7 from improvement list)
    partial: bool = False  # True if agent produced partial results


@dataclass
class DegradationInfo:
    """Structured metadata about missing/failed agents for the frontend."""

    missing_agents: list[str] = field(default_factory=list)
    reasons: dict[str, str] = field(default_factory=dict)  # agent_name → reason
    confidence_adjustment: int = 0  # Negative offset to apply to confidence scores


class CitationCollector:
    """Collects citations from agent computations for cross-referencing (Req 4.8).

    Each agent populates this as it produces insights, linking claims to
    source data points from the data platform.
    """

    def __init__(self) -> None:
        self._citations: list[Citation] = []

    def add(
        self,
        source: str,
        claim: str,
        data_point: str,
        url: str = "",
    ) -> None:
        """Record a citation linking a claim to its source data."""
        self._citations.append(Citation(
            source=source,
            claim=claim,
            data_point=data_point,
            url=url,
        ))

    @property
    def citations(self) -> list[Citation]:
        return list(self._citations)

    def clear(self) -> None:
        self._citations.clear()


class BaseAgent(ABC, Generic[T]):
    """Abstract base class for all agents in the multi-agent system.

    Subclasses implement `_run()` with their specific logic. The base class
    handles timeout enforcement, error wrapping, timing, and logging.
    """

    def __init__(self, name: str, timeout: float = DEFAULT_TIMEOUT_SECONDS) -> None:
        self.name = name
        self.timeout = timeout
        self.citation_collector = CitationCollector()

    async def run(self, ctx: AgentContext) -> AgentResult[T]:
        """Execute the agent with timeout enforcement and error handling.

        This is the public entry point. It wraps `_run()` with:
        - asyncio timeout (default 30s per Req 4.6)
        - Elapsed time tracking
        - Structured error capture (no exceptions leak out)
        - Citation collection
        """
        start = time.monotonic()
        self.citation_collector.clear()

        try:
            output = await asyncio.wait_for(
                self._run(ctx),
                timeout=self.timeout,
            )
            elapsed = time.monotonic() - start
            logger.info(
                "Agent %s completed for %s in %.2fs",
                self.name, ctx.symbol, elapsed,
            )
            return AgentResult(
                agent_name=self.name,
                output=output,
                citations=self.citation_collector.citations,
                elapsed_seconds=elapsed,
                success=True,
            )
        except asyncio.TimeoutError:
            elapsed = time.monotonic() - start
            logger.warning(
                "Agent %s timed out for %s after %.1fs",
                self.name, ctx.symbol, elapsed,
            )
            return AgentResult(
                agent_name=self.name,
                elapsed_seconds=elapsed,
                success=False,
                error=f"Timeout after {self.timeout}s",
            )
        except Exception as exc:
            elapsed = time.monotonic() - start
            logger.exception(
                "Agent %s failed for %s after %.2fs: %s",
                self.name, ctx.symbol, elapsed, exc,
            )
            return AgentResult(
                agent_name=self.name,
                elapsed_seconds=elapsed,
                success=False,
                error=str(exc),
            )

    @abstractmethod
    async def _run(self, ctx: AgentContext) -> T:
        """Implement agent-specific logic. Called by `run()` with timeout."""
        ...
