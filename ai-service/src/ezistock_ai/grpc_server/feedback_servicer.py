"""gRPC servicer for ezistock.FeedbackService.

Delegates to the Feedback Loop Engine for agent accuracy tracking
and model performance monitoring.

Requirements: 1.2, 43.4
"""

from __future__ import annotations

import logging
from datetime import datetime, timezone

import grpc

from ezistock_ai.generated.proto import feedback_pb2, feedback_pb2_grpc

logger = logging.getLogger(__name__)


class FeedbackServicer(feedback_pb2_grpc.FeedbackServiceServicer):
    """Concrete FeedbackService implementation.

    Each RPC delegates to the Feedback Loop Engine which will be wired
    in task 19.x.  Until then, methods return placeholder responses so
    the gRPC server can start and respond to health probes.
    """

    def GetAgentAccuracy(
        self,
        request: feedback_pb2.AccuracyRequest,
        context: grpc.ServicerContext,
    ) -> feedback_pb2.AccuracyResponse:
        """Return per-agent accuracy scores and bias detection."""
        logger.info(
            "GetAgentAccuracy called agent=%s period=%s",
            request.agent_name,
            request.period,
        )

        # TODO(task-19): delegate to feedback_engine.get_accuracy(request)
        return feedback_pb2.AccuracyResponse(
            agents=[],
            overall_accuracy=0.0,
            computed_at=datetime.now(timezone.utc).isoformat(),
        )

    def GetModelPerformance(
        self,
        request: feedback_pb2.ModelPerfRequest,
        context: grpc.ServicerContext,
    ) -> feedback_pb2.ModelPerfResponse:
        """Return Alpha Mining model performance metrics."""
        logger.info(
            "GetModelPerformance called model=%s period=%s",
            request.model_name,
            request.period,
        )

        # TODO(task-19): delegate to feedback_engine.get_model_perf(request)
        return feedback_pb2.ModelPerfResponse(
            models=[],
            computed_at=datetime.now(timezone.utc).isoformat(),
        )
