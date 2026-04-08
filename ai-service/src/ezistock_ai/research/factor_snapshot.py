"""Weekly Factor Snapshot report.

Compares factor group performance (Foreign, Value, Momentum, Quality, Low Vol)
against VNINDEX benchmark over trailing periods.

Requirements: 35.1
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field

import numpy as np
import pandas as pd

logger = logging.getLogger(__name__)

FACTOR_GROUPS = ["Foreign", "Value", "Momentum", "Quality", "LowVol"]


@dataclass
class FactorPerformance:
    """Performance of a single factor group."""

    name: str
    return_1w: float = 0.0
    return_1m: float = 0.0
    return_3m: float = 0.0
    return_ytd: float = 0.0
    sharpe_3m: float = 0.0
    vs_vnindex_1m: float = 0.0  # Excess return vs VNINDEX


@dataclass
class FactorSnapshotData:
    """Data for the weekly factor snapshot report."""

    report_date: str
    factors: list[FactorPerformance] = field(default_factory=list)
    vnindex_return_1w: float = 0.0
    vnindex_return_1m: float = 0.0
    top_factor: str = ""
    worst_factor: str = ""
    commentary: str = ""


class FactorSnapshotGenerator:
    """Generates weekly factor snapshot from signal space data."""

    def generate(
        self,
        signal_space: pd.DataFrame,
        forward_returns: pd.Series,
        vnindex_returns: pd.Series,
        report_date: str,
    ) -> FactorSnapshotData:
        """Build factor snapshot data for the given date.

        Args:
            signal_space: Signal space from DataLayer.
            forward_returns: Per-symbol daily returns.
            vnindex_returns: VN-Index daily returns.
            report_date: ISO date string for the report.
        """
        snapshot = FactorSnapshotData(report_date=report_date)

        if signal_space.empty or forward_returns.empty:
            return snapshot

        # Compute VNINDEX trailing returns
        if not vnindex_returns.empty:
            snapshot.vnindex_return_1w = float((1 + vnindex_returns.tail(5)).prod() - 1)
            snapshot.vnindex_return_1m = float((1 + vnindex_returns.tail(21)).prod() - 1)

        # For each factor group, compute top-quintile portfolio returns
        factor_column_map = {
            "Foreign": "flow_foreign_net",
            "Value": "fund_pb",
            "Momentum": "price_ret_60d",
            "Quality": "fund_roe",
            "LowVol": "tech_atr_14",
        }

        for group_name, col in factor_column_map.items():
            perf = self._compute_factor_return(
                signal_space, forward_returns, col, group_name,
                invert=(group_name == "LowVol"),  # Low vol = sort ascending
            )
            perf.vs_vnindex_1m = perf.return_1m - snapshot.vnindex_return_1m
            snapshot.factors.append(perf)

        # Identify top/worst
        if snapshot.factors:
            best = max(snapshot.factors, key=lambda f: f.return_1m)
            worst = min(snapshot.factors, key=lambda f: f.return_1m)
            snapshot.top_factor = best.name
            snapshot.worst_factor = worst.name

        return snapshot

    @staticmethod
    def _compute_factor_return(
        signal_space: pd.DataFrame,
        returns: pd.Series,
        factor_col: str,
        group_name: str,
        invert: bool = False,
    ) -> FactorPerformance:
        """Compute top-quintile portfolio return for a factor."""
        perf = FactorPerformance(name=group_name)

        if factor_col not in signal_space.columns:
            return perf

        dates = signal_space.index.get_level_values("date").unique().sort_values()
        if len(dates) < 5:
            return perf

        # Get latest date's factor values
        latest = dates[-1]
        day_signals = signal_space.xs(latest, level="date")
        if factor_col not in day_signals.columns:
            return perf

        factor_vals = day_signals[factor_col].dropna()
        if len(factor_vals) < 10:
            return perf

        # Top quintile
        if invert:
            top_symbols = factor_vals.nsmallest(len(factor_vals) // 5).index
        else:
            top_symbols = factor_vals.nlargest(len(factor_vals) // 5).index

        # Compute trailing returns for the top-quintile portfolio
        for n_days, attr in [(5, "return_1w"), (21, "return_1m"), (63, "return_3m")]:
            if len(dates) >= n_days:
                period_dates = dates[-n_days:]
                period_returns = returns.reindex(
                    pd.MultiIndex.from_product([period_dates, top_symbols], names=["date", "symbol"])
                ).fillna(0)
                if not period_returns.empty:
                    daily_avg = period_returns.groupby(level="date").mean()
                    setattr(perf, attr, float((1 + daily_avg).prod() - 1))

        return perf
