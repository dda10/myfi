"""Generated gRPC stubs for ezistock.AlphaMiningService.

Hand-written to match protoc-gen-grpc_python output for alpha.proto.
Replace with protoc-generated output when toolchain is available.
"""

from __future__ import annotations

import grpc

from ezistock_ai.generated.proto import alpha_pb2

# ---------------------------------------------------------------------------
# Service descriptor constants
# ---------------------------------------------------------------------------

_ALPHA_SERVICE_NAME = "ezistock.AlphaMiningService"

_GET_RANKING_METHOD = f"/{_ALPHA_SERVICE_NAME}/GetRanking"
_RUN_BACKTEST_METHOD = f"/{_ALPHA_SERVICE_NAME}/RunBacktest"
_GET_REGIME_METHOD = f"/{_ALPHA_SERVICE_NAME}/GetRegime"


# ---------------------------------------------------------------------------
# Servicer (server-side abstract base)
# ---------------------------------------------------------------------------


class AlphaMiningServiceServicer:
    """Base class for AlphaMiningService server implementations."""

    def GetRanking(
        self,
        request: alpha_pb2.RankingRequest,
        context: grpc.ServicerContext,
    ) -> alpha_pb2.RankingResponse:
        """Return consensus-based stock rankings from the strategy ensemble."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method GetRanking not implemented")
        raise NotImplementedError("Method GetRanking not implemented")

    def RunBacktest(
        self,
        request: alpha_pb2.BacktestRequest,
        context: grpc.ServicerContext,
    ) -> alpha_pb2.BacktestResponse:
        """Execute walk-forward backtests with regime-aware validation."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method RunBacktest not implemented")
        raise NotImplementedError("Method RunBacktest not implemented")

    def GetRegime(
        self,
        request: alpha_pb2.RegimeRequest,
        context: grpc.ServicerContext,
    ) -> alpha_pb2.RegimeResponse:
        """Return the current market regime classification."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method GetRegime not implemented")
        raise NotImplementedError("Method GetRegime not implemented")


# ---------------------------------------------------------------------------
# Server registration
# ---------------------------------------------------------------------------

_ALPHA_SERVICE_HANDLER_METHODS = {
    "GetRanking": (
        alpha_pb2.RankingRequest,
        alpha_pb2.RankingResponse,
    ),
    "RunBacktest": (
        alpha_pb2.BacktestRequest,
        alpha_pb2.BacktestResponse,
    ),
    "GetRegime": (
        alpha_pb2.RegimeRequest,
        alpha_pb2.RegimeResponse,
    ),
}


def _make_handler(servicer: AlphaMiningServiceServicer, method_name: str):
    """Create a unary-unary handler for the given method."""
    method = getattr(servicer, method_name)

    def handler(request, context):
        return method(request, context)

    return grpc.unary_unary_rpc_method_handler(handler)


def add_AlphaMiningServiceServicer_to_server(
    servicer: AlphaMiningServiceServicer,
    server: grpc.Server,
) -> None:
    """Register an AlphaMiningServiceServicer with a gRPC server."""
    rpc_method_handlers = {
        name: _make_handler(servicer, name)
        for name in _ALPHA_SERVICE_HANDLER_METHODS
    }
    generic_handler = grpc.method_service_handler(
        _ALPHA_SERVICE_NAME,
        rpc_method_handlers,
    )
    server.add_generic_rpc_handlers((generic_handler,))
