"""Generated gRPC stubs for ezistock.FeedbackService.

Hand-written to match protoc-gen-grpc_python output for feedback.proto.
Replace with protoc-generated output when toolchain is available.
"""

from __future__ import annotations

import grpc

from ezistock_ai.generated.proto import feedback_pb2

# ---------------------------------------------------------------------------
# Service descriptor constants
# ---------------------------------------------------------------------------

_FEEDBACK_SERVICE_NAME = "ezistock.FeedbackService"

_GET_AGENT_ACCURACY_METHOD = f"/{_FEEDBACK_SERVICE_NAME}/GetAgentAccuracy"
_GET_MODEL_PERFORMANCE_METHOD = f"/{_FEEDBACK_SERVICE_NAME}/GetModelPerformance"


# ---------------------------------------------------------------------------
# Servicer (server-side abstract base)
# ---------------------------------------------------------------------------


class FeedbackServiceServicer:
    """Base class for FeedbackService server implementations."""

    def GetAgentAccuracy(
        self,
        request: feedback_pb2.AccuracyRequest,
        context: grpc.ServicerContext,
    ) -> feedback_pb2.AccuracyResponse:
        """Return per-agent accuracy scores and bias detection."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method GetAgentAccuracy not implemented")
        raise NotImplementedError("Method GetAgentAccuracy not implemented")

    def GetModelPerformance(
        self,
        request: feedback_pb2.ModelPerfRequest,
        context: grpc.ServicerContext,
    ) -> feedback_pb2.ModelPerfResponse:
        """Return Alpha Mining model performance metrics."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method GetModelPerformance not implemented")
        raise NotImplementedError("Method GetModelPerformance not implemented")


# ---------------------------------------------------------------------------
# Server registration
# ---------------------------------------------------------------------------

_FEEDBACK_SERVICE_HANDLER_METHODS = {
    "GetAgentAccuracy": (
        feedback_pb2.AccuracyRequest,
        feedback_pb2.AccuracyResponse,
    ),
    "GetModelPerformance": (
        feedback_pb2.ModelPerfRequest,
        feedback_pb2.ModelPerfResponse,
    ),
}


def _make_handler(servicer: FeedbackServiceServicer, method_name: str):
    """Create a unary-unary handler for the given method."""
    method = getattr(servicer, method_name)

    def handler(request, context):
        return method(request, context)

    return grpc.unary_unary_rpc_method_handler(handler)


def add_FeedbackServiceServicer_to_server(
    servicer: FeedbackServiceServicer,
    server: grpc.Server,
) -> None:
    """Register a FeedbackServiceServicer with a gRPC server."""
    rpc_method_handlers = {
        name: _make_handler(servicer, name)
        for name in _FEEDBACK_SERVICE_HANDLER_METHODS
    }
    generic_handler = grpc.method_service_handler(
        _FEEDBACK_SERVICE_NAME,
        rpc_method_handlers,
    )
    server.add_generic_rpc_handlers((generic_handler,))
