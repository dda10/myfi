"""gRPC servicer for ezistock.AgentService.

Delegates to the multi-agent orchestrator for stock analysis, investment ideas,
chat, and hot topics detection.

Requirements: 1.2, 43.4
"""

from __future__ import annotations

import asyncio
import logging
from datetime import datetime, timezone

import grpc

from ezistock_ai.agents.orchestrator import AgentOrchestrator
from ezistock_ai.generated.proto import agent_pb2, agent_pb2_grpc
from ezistock_ai.llm.router import LLMRouter

logger = logging.getLogger(__name__)


class AgentServicer(agent_pb2_grpc.AgentServiceServicer):
    """Concrete AgentService implementation.

    Delegates AnalyzeStock to the multi-agent orchestrator pipeline.
    Chat, Ideas, and HotTopics are wired to the orchestrator when
    those features are implemented (tasks 20.3, etc.).
    """

    def __init__(self, llm_router: LLMRouter | None = None) -> None:
        self._orchestrator = AgentOrchestrator(llm_router=llm_router)

    def AnalyzeStock(
        self,
        request: agent_pb2.AnalyzeStockRequest,
        context: grpc.ServicerContext,
    ) -> agent_pb2.AnalyzeStockResponse:
        """Run the full multi-agent pipeline for a single stock."""
        logger.info("AnalyzeStock called for symbol=%s", request.symbol)

        try:
            loop = asyncio.get_event_loop()
            if loop.is_running():
                import concurrent.futures
                with concurrent.futures.ThreadPoolExecutor() as pool:
                    result = pool.submit(
                        asyncio.run,
                        self._orchestrator.analyze(request),
                    ).result(timeout=90)
            else:
                result = asyncio.run(self._orchestrator.analyze(request))

            return self._orchestrator.to_protobuf(result)

        except Exception as exc:
            logger.exception("AnalyzeStock failed for %s: %s", request.symbol, exc)
            return agent_pb2.AnalyzeStockResponse(
                technical=agent_pb2.TechnicalAnalysis(
                    symbol=request.symbol,
                    composite_signal="neutral",
                ),
                disclaimer=f"Analysis failed: {exc}",
            )

    def GenerateInvestmentIdeas(
        self,
        request: agent_pb2.IdeaRequest,
        context: grpc.ServicerContext,
    ) -> agent_pb2.IdeaResponse:
        """Generate proactive buy/sell investment ideas."""
        logger.info(
            "GenerateInvestmentIdeas called for user=%s, max_ideas=%d",
            request.user_id,
            request.max_ideas,
        )
        # TODO(task-20.3): Wire to orchestrator idea generation
        return agent_pb2.IdeaResponse(
            ideas=[],
            generated_at=datetime.now(timezone.utc).isoformat(),
        )

    def Chat(
        self,
        request: agent_pb2.ChatRequest,
        context: grpc.ServicerContext,
    ) -> agent_pb2.ChatResponse:
        """Handle conversational AI interaction."""
        logger.info("Chat called for user=%s", request.user_id)
        # TODO: Wire to orchestrator chat with memory
        return agent_pb2.ChatResponse(
            response="Xin chào! Hệ thống AI đang được triển khai. Vui lòng thử lại sau.",
            citations=[],
            suggestions=[],
            disclaimer="This is AI-generated content, not financial advice.",
        )

    def GetHotTopics(
        self,
        request: agent_pb2.HotTopicsRequest,
        context: grpc.ServicerContext,
    ) -> agent_pb2.HotTopicsResponse:
        """Detect trending stocks and market events."""
        logger.info(
            "GetHotTopics called limit=%d market=%s",
            request.limit,
            request.market,
        )
        # TODO: Wire to orchestrator hot topics detection
        return agent_pb2.HotTopicsResponse(
            topics=[],
            generated_at=datetime.now(timezone.utc).isoformat(),
        )
