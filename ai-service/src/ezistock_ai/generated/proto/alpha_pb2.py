"""Generated protobuf message classes for ezistock/alpha.proto.

Hand-written stubs matching the proto definitions.
Replace with protoc-generated output when toolchain is available.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Dict, List, Optional


# ---------------------------------------------------------------------------
# GetRanking
# ---------------------------------------------------------------------------


@dataclass
class FactorWeight:
    factor: str = ""
    weight: float = 0.0


@dataclass
class RankingRequest:
    universe: List[str] = field(default_factory=list)
    factor_groups: List[str] = field(default_factory=list)
    custom_weights: List[FactorWeight] = field(default_factory=list)
    top_n: int = 0
    rebalance_frequency: str = ""


@dataclass
class RankedStock:
    symbol: str = ""
    composite_score: float = 0.0
    rank: int = 0
    factor_scores: Dict[str, float] = field(default_factory=dict)
    strategy_agreement: int = 0
    liquidity_tier: str = ""


@dataclass
class RankingResponse:
    rankings: List[RankedStock] = field(default_factory=list)
    regime: str = ""
    generated_at: str = ""
    total_universe_size: int = 0


# ---------------------------------------------------------------------------
# RunBacktest
# ---------------------------------------------------------------------------


@dataclass
class BacktestRequest:
    universe: List[str] = field(default_factory=list)
    factor_groups: List[str] = field(default_factory=list)
    custom_weights: List[FactorWeight] = field(default_factory=list)
    start_date: str = ""
    end_date: str = ""
    top_n: int = 0
    rebalance_frequency: str = ""
    commission_rate: float = 0.0


@dataclass
class MonthlyReturn:
    month: str = ""
    return_pct: float = 0.0


@dataclass
class DrawdownPeriod:
    start_date: str = ""
    end_date: str = ""
    max_drawdown: float = 0.0


@dataclass
class BacktestMetrics:
    cumulative_return: float = 0.0
    annualized_return: float = 0.0
    sharpe_ratio: float = 0.0
    max_drawdown: float = 0.0
    win_rate: float = 0.0
    profit_factor: float = 0.0
    information_ratio: float = 0.0
    total_trades: int = 0


@dataclass
class BacktestResponse:
    metrics: Optional[BacktestMetrics] = None
    benchmark_metrics: Optional[BacktestMetrics] = None
    monthly_returns: List[MonthlyReturn] = field(default_factory=list)
    drawdowns: List[DrawdownPeriod] = field(default_factory=list)
    alpha_decay_warnings: List[str] = field(default_factory=list)
    regime_during_test: str = ""


# ---------------------------------------------------------------------------
# GetRegime
# ---------------------------------------------------------------------------


@dataclass
class RegimeRequest:
    pass


@dataclass
class RegimeIndicator:
    name: str = ""
    value: float = 0.0
    signal: str = ""


@dataclass
class RegimeResponse:
    regime: str = ""
    confidence: float = 0.0
    indicators: List[RegimeIndicator] = field(default_factory=list)
    detected_at: str = ""
