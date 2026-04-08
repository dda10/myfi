"""Monthly Sector Deep-Dive report.

Analyzes each ICB sector's performance, fundamentals, and top picks.

Requirements: 35.5
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field

import pandas as pd

logger = logging.getLogger(__name__)


@dataclass
class SectorPick:
    """A top stock pick within a sector."""

    symbol: str
    score: float
    reasoning: str = ""


@dataclass
class SectorAnalysis:
    """Analysis for a single sector."""

    sector_code: str
    sector_name: str
    return_1m: float = 0.0
    return_3m: float = 0.0
    median_pe: float = 0.0
    median_pb: float = 0.0
    foreign_flow_trend: str = "neutral"
    top_picks: list[SectorPick] = field(default_factory=list)
    outlook: str = ""


@dataclass
class SectorDeepDiveData:
    """Data for the monthly sector deep-dive report."""

    report_date: str
    sectors: list[SectorAnalysis] = field(default_factory=list)
    top_sector: str = ""
    worst_sector: str = ""


class SectorDeepDiveGenerator:
    """Generates monthly sector deep-dive from market data."""

    def generate(
        self,
        sector_performance: list[dict],
        signal_space: pd.DataFrame,
        report_date: str,
    ) -> SectorDeepDiveData:
        """Build sector deep-dive data.

        Args:
            sector_performance: List of dicts with sector_code, sector_name, return_1m, return_3m, etc.
            signal_space: Signal space for stock-level analysis.
            report_date: ISO date string.
        """
        data = SectorDeepDiveData(report_date=report_date)

        for sp in sector_performance:
            analysis = SectorAnalysis(
                sector_code=sp.get("sector_code", ""),
                sector_name=sp.get("sector_name", ""),
                return_1m=sp.get("return_1m", 0.0),
                return_3m=sp.get("return_3m", 0.0),
                median_pe=sp.get("median_pe", 0.0),
                median_pb=sp.get("median_pb", 0.0),
                foreign_flow_trend=sp.get("foreign_flow_trend", "neutral"),
            )

            # Find top picks in this sector from signal space
            sector_symbols = sp.get("symbols", [])
            if sector_symbols and not signal_space.empty:
                analysis.top_picks = self._find_top_picks(signal_space, sector_symbols)

            data.sectors.append(analysis)

        if data.sectors:
            best = max(data.sectors, key=lambda s: s.return_1m)
            worst = min(data.sectors, key=lambda s: s.return_1m)
            data.top_sector = best.sector_name
            data.worst_sector = worst.sector_name

        return data

    @staticmethod
    def _find_top_picks(signal_space: pd.DataFrame, symbols: list[str], n: int = 3) -> list[SectorPick]:
        """Find top N stocks in a sector by composite signal score."""
        dates = signal_space.index.get_level_values("date").unique()
        if len(dates) == 0:
            return []

        latest = dates[-1]
        try:
            day_data = signal_space.xs(latest, level="date")
        except KeyError:
            return []

        sector_data = day_data[day_data.index.isin(symbols)]
        if sector_data.empty:
            return []

        # Simple composite: mean of all numeric columns
        scores = sector_data.select_dtypes(include=["number"]).mean(axis=1)
        top = scores.nlargest(n)

        return [
            SectorPick(symbol=str(sym), score=float(score))
            for sym, score in top.items()
        ]
