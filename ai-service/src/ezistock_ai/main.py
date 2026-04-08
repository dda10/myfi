"""Entry point — starts gRPC server (port 50051) and FastAPI REST fallback (port 8000)."""

import asyncio
import logging
import signal
from concurrent import futures

import grpc
import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from ezistock_ai.config import Config

logger = logging.getLogger("ezistock_ai")

config = Config()

# ---------------------------------------------------------------------------
# FastAPI REST fallback
# ---------------------------------------------------------------------------

app = FastAPI(
    title="EziStock AI Service",
    version="0.1.0",
    description="Multi-agent AI system and Alpha Mining engine for Vietnamese stock analysis",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
async def health() -> dict[str, str]:
    """Health check endpoint."""
    return {"status": "ok", "service": "ezistock-ai"}


# ---------------------------------------------------------------------------
# gRPC server
# ---------------------------------------------------------------------------


def _create_grpc_server() -> grpc.aio.Server:
    """Create and configure the async gRPC server with all servicers registered."""
    from ezistock_ai.generated.proto import (
        agent_pb2_grpc,
        alpha_pb2_grpc,
        feedback_pb2_grpc,
    )
    from ezistock_ai.grpc_server.agent_servicer import AgentServicer
    from ezistock_ai.grpc_server.alpha_servicer import AlphaMiningServicer
    from ezistock_ai.grpc_server.feedback_servicer import FeedbackServicer
    from ezistock_ai.llm.router import LLMRouter

    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))

    llm_router = LLMRouter(config)
    agent_pb2_grpc.add_AgentServiceServicer_to_server(AgentServicer(llm_router=llm_router), server)
    alpha_pb2_grpc.add_AlphaMiningServiceServicer_to_server(AlphaMiningServicer(), server)
    feedback_pb2_grpc.add_FeedbackServiceServicer_to_server(FeedbackServicer(), server)

    server.add_insecure_port(f"[::]:{config.grpc_port}")
    return server


# ---------------------------------------------------------------------------
# Lifecycle
# ---------------------------------------------------------------------------


async def _serve() -> None:
    """Start both gRPC and FastAPI servers concurrently."""
    logging.basicConfig(
        level=logging.DEBUG if config.debug else logging.INFO,
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    )

    # gRPC
    grpc_server = _create_grpc_server()
    await grpc_server.start()
    logger.info("gRPC server listening on port %d", config.grpc_port)

    # FastAPI (uvicorn)
    uvi_config = uvicorn.Config(
        app,
        host="0.0.0.0",
        port=config.rest_port,
        log_level="debug" if config.debug else "info",
    )
    uvi_server = uvicorn.Server(uvi_config)

    # Graceful shutdown on SIGINT / SIGTERM
    stop_event = asyncio.Event()

    def _signal_handler() -> None:
        logger.info("Shutdown signal received")
        stop_event.set()

    loop = asyncio.get_running_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, _signal_handler)

    logger.info("FastAPI REST fallback listening on port %d", config.rest_port)

    # Run both servers; stop when either finishes or signal received
    uvi_task = asyncio.create_task(uvi_server.serve())
    stop_task = asyncio.create_task(stop_event.wait())

    await asyncio.wait(
        [uvi_task, stop_task],
        return_when=asyncio.FIRST_COMPLETED,
    )

    # Cleanup
    logger.info("Shutting down servers…")
    await grpc_server.stop(grace=5)
    uvi_server.should_exit = True
    await uvi_task
    logger.info("Shutdown complete")


def main() -> None:
    """CLI entry point."""
    asyncio.run(_serve())


if __name__ == "__main__":
    main()
