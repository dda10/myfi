"""Generated gRPC stubs for ezistock.AgentService.

Hand-written to match protoc-gen-grpc_python output for agent.proto.
Replace with protoc-generated output when toolchain is available.
"""

from __future__ import annotations

import grpc

from ezistock_ai.generated.proto import agent_pb2

# ---------------------------------------------------------------------------
# Service descriptor constants
# ---------------------------------------------------------------------------

_AGENT_SERVICE_NAME = "ezistock.AgentService"

_ANALYZE_STOCK_METHOD = f"/{_AGENT_SERVICE_NAME}/AnalyzeStock"
_GENERATE_IDEAS_METHOD = f"/{_AGENT_SERVICE_NAME}/GenerateInvestmentIdeas"
_CHAT_METHOD = f"/{_AGENT_SERVICE_NAME}/Chat"
_HOT_TOPICS_METHOD = f"/{_AGENT_SERVICE_NAME}/GetHotTopics"


# ---------------------------------------------------------------------------
# Servicer (server-side abstract base)
# ---------------------------------------------------------------------------


class AgentServiceServicer:
    """Base class for AgentService server implementations."""

    def AnalyzeStock(
        self,
        request: agent_pb2.AnalyzeStockRequest,
        context: grpc.ServicerContext,
    ) -> agent_pb2.AnalyzeStockResponse:
        """Perform comprehensive stock analysis using the multi-agent system."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method AnalyzeStock not implemented")
        raise NotImplementedError("Method AnalyzeStock not implemented")

    def GenerateInvestmentIdeas(
        self,
        request: agent_pb2.IdeaRequest,
        context: grpc.ServicerContext,
    ) -> agent_pb2.IdeaResponse:
        """Produce proactive buy/sell recommendations."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method GenerateInvestmentIdeas not implemented")
        raise NotImplementedError("Method GenerateInvestmentIdeas not implemented")

    def Chat(
        self,
        request: agent_pb2.ChatRequest,
        context: grpc.ServicerContext,
    ) -> agent_pb2.ChatResponse:
        """Handle conversational AI interactions with citation support."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method Chat not implemented")
        raise NotImplementedError("Method Chat not implemented")

    def GetHotTopics(
        self,
        request: agent_pb2.HotTopicsRequest,
        context: grpc.ServicerContext,
    ) -> agent_pb2.HotTopicsResponse:
        """Detect trending stocks and market events."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method GetHotTopics not implemented")
        raise NotImplementedError("Method GetHotTopics not implemented")


# ---------------------------------------------------------------------------
# Server registration
# ---------------------------------------------------------------------------

# Method handler descriptors used by add_AgentServiceServicer_to_server.
_AGENT_SERVICE_HANDLER_METHODS = {
    "AnalyzeStock": (
        agent_pb2.AnalyzeStockRequest,
        agent_pb2.AnalyzeStockResponse,
    ),
    "GenerateInvestmentIdeas": (
        agent_pb2.IdeaRequest,
        agent_pb2.IdeaResponse,
    ),
    "Chat": (
        agent_pb2.ChatRequest,
        agent_pb2.ChatResponse,
    ),
    "GetHotTopics": (
        agent_pb2.HotTopicsRequest,
        agent_pb2.HotTopicsResponse,
    ),
}


def _make_handler(servicer: AgentServiceServicer, method_name: str):
    """Create a unary-unary handler for the given method."""
    method = getattr(servicer, method_name)

    def handler(request, context):
        return method(request, context)

    return grpc.unary_unary_rpc_method_handler(handler)


def add_AgentServiceServicer_to_server(
    servicer: AgentServiceServicer,
    server: grpc.Server,
) -> None:
    """Register an AgentServiceServicer with a gRPC server."""
    rpc_method_handlers = {
        name: _make_handler(servicer, name)
        for name in _AGENT_SERVICE_HANDLER_METHODS
    }
    generic_handler = grpc.method_service_handler(
        _AGENT_SERVICE_NAME,
        rpc_method_handlers,
    )
    server.add_generic_rpc_handlers((generic_handler,))
