"""Generated protobuf message classes for ezistock/feedback.proto.

Hand-written stubs matching the proto definitions.
Replace with protoc-generated output when toolchain is available.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Dict, List, Optional


# ---------------------------------------------------------------------------
# GetAgentAccuracy
# ---------------------------------------------------------------------------


@dataclass
class AccuracyRequest:
    agent_name: str = ""
    period: str = ""


@dataclass
class BiasIndicator:
    type: str = ""
    magnitude: float = 0.0
    description: str = ""


@dataclass
class AgentAccuracy:
    agent_name: str = ""
    accuracy_percent: float = 0.0
    total_predictions: int = 0
    correct_predictions: int = 0
    biases: List[BiasIndicator] = field(default_factory=list)
    period: str = ""


@dataclass
class AccuracyResponse:
    agents: List[AgentAccuracy] = field(default_factory=list)
    overall_accuracy: float = 0.0
    computed_at: str = ""


# ---------------------------------------------------------------------------
# GetModelPerformance
# ---------------------------------------------------------------------------


@dataclass
class ModelPerfRequest:
    model_name: str = ""
    period: str = ""


@dataclass
class ModelPerformance:
    model_name: str = ""
    model_version: str = ""
    accuracy: float = 0.0
    precision_score: float = 0.0
    recall: float = 0.0
    f1_score: float = 0.0
    feature_importance: Dict[str, float] = field(default_factory=dict)
    trained_at: str = ""
    regime_context: str = ""
    alpha_decay_detected: bool = False


@dataclass
class ModelPerfResponse:
    models: List[ModelPerformance] = field(default_factory=list)
    computed_at: str = ""
