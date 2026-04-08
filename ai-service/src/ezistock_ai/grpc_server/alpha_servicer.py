"""gRPC servicer for ezistock.AlphaMiningService.

Delegates to the Alpha Mining Engine for stock ranking, backtesting,
and market regime detection.

Requirements: 1.2, 43.4
"""

from __future__ import annotations

import logging
from datetime import datetime, timezone

import grpc

from ezistock_ai.generated.proto import alpha_pb2, alpha_pb2_grpc

logger = logging.getLogger(__name__)


class AlphaMiningServicer(alpha_pb2_grpc.AlphaMiningServiceServicer):
    """Concrete AlphaMiningService implementation.

    Each RPC delegates to the Alpha Mining Engine layers which will be
    wired in tasks 18.x.  Until then, methods return placeholder responses
    so the gRPC server can start and respond to health probes.
    """

    def GetRanking(
        self,
        request: alpha_pb2.RankingRequest,
        context: grpc.ServicerContext,
    ) -> alpha_pb2.RankingResponse:
        """Return consensus-based stock rankings."""
        logger.info(
            "GetRanking called universe=%s factor_groups=%s top_n=%d",
            request.universe,
            request.factor_groups,
            request.top_n,
        )

        # TODO(task-18): delegate to deployment_layer.get_ranking(request)
        return alpha_pb2.RankingResponse(
            rankings=[],
            regime="sideways",
            generated_at=datetime.now(timezone.utc).isoformat(),
            total_universe_size=0,
        )

    def RunBacktest(
        self,
        request: alpha_pb2.BacktestRequest,
        context: grpc.ServicerContext,
    ) -> alpha_pb2.BacktestResponse:
        """Execute walk-forward backtests with regime-aware validation."""
        logger.info(
            "RunBacktest called universe=%s start=%s end=%s",
            request.universe,
            request.start_date,
            request.end_date,
        )

        # TODO(task-18): delegate to backtest_layer.run(request)
        return alpha_pb2.BacktestResponse(
            metrics=alpha_pb2.BacktestMetrics(),
            benchmark_metrics=alpha_pb2.BacktestMetrics(),
            monthly_returns=[],
            drawdowns=[],
            alpha_decay_warnings=[],
            regime_during_test="sideways",
        )

    def GetRegime(
        self,
        request: alpha_pb2.RegimeRequest,
        context: grpc.ServicerContext,
    ) -> alpha_pb2.RegimeResponse:
        """Return the current market regime classification."""
        logger.info("GetRegime called")

        # TODO(task-18): delegate to regime_detector.detect()
        return alpha_pb2.RegimeResponse(
            regime="sideways",
            confidence=0.0,
            indicators=[],
            detected_at=datetime.now(timezone.utc).isoformat(),
        )
